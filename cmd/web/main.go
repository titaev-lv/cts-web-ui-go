// Package main - точка входа в приложение.
// Этот файл запускается первым при старте программы.
package main

import (
	"context"
	"ctweb/internal/config"      // Пакет для работы с конфигурацией
	"ctweb/internal/controllers" // Контроллеры (обработчики HTTP запросов)
	"ctweb/internal/db"          // Подключение к базе данных
	"ctweb/internal/logger"      // Система логирования
	"ctweb/internal/middleware"  // Middleware (промежуточные обработчики)
	"ctweb/internal/session"     // Управление сессиями
	"fmt"                        // Форматирование строк
	"html/template"              // HTML шаблоны
	"net/http"                   // HTTP клиент и сервер
	"os"                         // Работа с операционной системой
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin" // Веб-фреймворк Gin
	"golang.org/x/net/http2"
)

// main - главная функция, которая выполняется при запуске программы.
// Здесь происходит инициализация всех компонентов приложения:
//  1. Загрузка конфигурации
//  2. Настройка режима работы
//  3. Подключение к базе данных
//  4. Настройка маршрутов (routes)
//  5. Запуск HTTP сервера
func main() {
	// ============================================
	// ШАГ 1: Загрузка конфигурации
	// ============================================
	// Загружаем конфигурацию из файла config/config.yaml
	// Пустая строка "" означает, что нужно искать в стандартных местах
	cfg, err := config.Load("")
	if err != nil {
		// Если не удалось загрузить конфигурацию, используем стандартный вывод
		// (логгер ещё не инициализирован, т.к. нужна конфигурация)
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// ============================================
	// ШАГ 2: Инициализация системы логирования
	// ============================================
	// Инициализируем логгер на основе конфигурации
	// После этого можно использовать logger.Info(), logger.Error() и т.д.
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()
	logger.Info().
		Str("app", cfg.App.Name).
		Str("version", cfg.App.Version).
		Msg("Application starting")

	// ============================================
	// ШАГ 2.1: Инициализация Session Manager
	// ============================================
	// Инициализируем систему управления сессиями
	// Session Manager использует cookies для хранения сессий и токенов "Remember Me"
	session.Init()
	logger.Info().Msg("Session manager initialized")

	// ============================================
	// ШАГ 3: Настройка режима работы Gin
	// ============================================
	// Режим Gin вычисляется от logging.level:
	//   - debug -> gin.DebugMode
	//   - иначе -> gin.ReleaseMode
	if cfg.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// ============================================
	// ШАГ 4: Подключение к базе данных
	// ============================================
	// Функция db.Connect() читает настройки БД из конфигурации
	// и устанавливает соединение с MySQL
	// Если подключение не удалось, программа завершится с ошибкой
	logger.Info().Msg("Connecting to database...")
	db.Connect()
	logger.Info().Msg("Database connected successfully")

	// ============================================
	// ШАГ 5: Настройка HTTP роутера (маршрутизатора)
	// ============================================
	// gin.New() создаёт роутер БЕЗ встроенных middleware
	// Мы используем свой Recovery middleware с интеграцией в нашу систему логирования
	r := gin.New()

	if cfg.Proxy.Enabled && cfg.Proxy.TrustForwardHeaders {
		if err := r.SetTrustedProxies(cfg.Proxy.TrustedCIDRs); err != nil {
			logger.Fatal().Err(err).Msg("Invalid proxy.trusted_cidrs")
		}
		r.ForwardedByClientIP = true
	} else {
		if err := r.SetTrustedProxies(nil); err != nil {
			logger.Fatal().Err(err).Msg("Failed to disable trusted proxies")
		}
		r.ForwardedByClientIP = false
	}

	// ============================================
	// ШАГ 5.1: Recovery Middleware (обработка паник)
	// ============================================
	// Recovery middleware должен быть ПЕРВЫМ, чтобы перехватывать паники
	// из всех последующих middleware и handlers.
	// Используем версию с опцией показа stack trace в режиме разработки
	showStack := cfg.IsDebug()
	r.Use(middleware.RecoveryMiddlewareWithStack(showStack))
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.AuditLogMiddleware())
	r.Use(middleware.AccessLogMiddleware())

	// ============================================
	// ШАГ 6: Регистрация Security Middleware
	// ============================================
	// Security middleware должен быть первым, чтобы защитить все запросы
	// SecurityHeadersMiddleware устанавливает HTTP заголовки безопасности
	// (X-Frame-Options, CSP, X-XSS-Protection и т.д.)
	r.Use(middleware.SecurityHeadersMiddleware())

	// ============================================
	// ШАГ 9: Регистрация контроллеров и маршрутов
	// ============================================
	// Создаём контроллеры
	userController := controllers.NewUserController()
	groupController := controllers.NewGroupController()
	exchangeController := controllers.NewExchangeController()
	exchangeAccountController := controllers.NewExchangeAccountController()

	// ============================================
	// ШАГ 8: Регистрация Auth Middleware
	// ============================================
	// Middleware - это функции, которые выполняются ДО обработки запроса
	// AuthMiddleware проверяет, авторизован ли пользователь
	// Если не авторизован - перенаправляет на страницу входа
	r.Use(middleware.AuthMiddleware())

	// Регистрируем маршруты (URL пути):
	//   GET /          -> главная страница (userController.Home)
	//   GET /login     -> страница входа (userController.ShowLoginPage)
	r.GET("/", userController.Home)
	r.GET("/login", userController.ShowLoginPage)

	// TEMP: Дебаг endpoint для проверки пользователя из контекста
	r.GET("/debug/user", func(c *gin.Context) {
		user, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":        user.ID,
			"login":     user.Login,
			"name":      user.Name,
			"last_name": user.LastName,
			"email":     user.Email,
			"active":    user.Active,
		})
	})

	// Создаём группу маршрутов для аутентификации
	// Группа позволяет применять middleware только к определённым маршрутам
	auth := r.Group("/auth")
	// POST /auth/login -> обработка формы входа (userController.Login)
	// Применяем rate limiting для защиты от брутфорса (5 попыток в минуту)
	auth.POST("/login", middleware.LoginRateLimitMiddleware(), userController.Login)
	// GET /auth/logout -> выход из системы (userController.Logout)
	// Требует авторизации (проверяется в AuthMiddleware)
	auth.GET("/logout", userController.Logout)

	// ============================================
	// ШАГ 9.1: Маршруты для управления пользователями (требуют авторизации)
	// ============================================
	// GET /users/ -> страница управления пользователями
	// POST /users/ajax_get_users -> получение списка пользователей для DataTables
	// GET /users/ajax_getid_user -> получение пользователя по ID
	// POST /users/ajax_create_user -> создание пользователя
	// POST /users/ajax_edit_user -> редактирование пользователя
	users := r.Group("/users")
	users.GET("/", userController.List)
	users.POST("/ajax_get_users", userController.AjaxGetUsers)
	users.GET("/ajax_getid_user", userController.AjaxGetUserById)
	users.POST("/ajax_getid_user", userController.AjaxGetUserById) // Поддержка POST тоже
	users.POST("/ajax_create_user", userController.AjaxCreateUser)
	users.POST("/ajax_edit_user", userController.AjaxEditUser)

	// ============================================
	// ШАГ 9.2: Маршруты для управления группами (требуют авторизации)
	// ============================================
	// GET /groups/ -> страница управления группами
	// POST /groups/ajax_get_groups -> получение списка групп для DataTables
	// GET /groups/ajax_getid_group -> получение группы по ID
	// POST /groups/ajax_create_group -> создание группы
	// POST /groups/ajax_edit_group -> редактирование группы
	groups := r.Group("/groups")
	groups.GET("/", groupController.List)
	groups.POST("/ajax_get_groups", groupController.AjaxGetGroups)
	groups.GET("/ajax_getid_group", groupController.AjaxGetGroupById)
	groups.POST("/ajax_getid_group", groupController.AjaxGetGroupById) // Поддержка POST тоже
	groups.POST("/ajax_create_group", groupController.AjaxCreateGroup)
	groups.POST("/ajax_edit_group", groupController.AjaxEditGroup)

	// ============================================
	// ШАГ 9.3: Маршруты для управления биржами (требуют авторизации, админ)
	// ============================================
	exchanges := r.Group("/exchange_manage")
	exchanges.GET("/", exchangeController.List)
	exchanges.POST("/ajax_get_exchanges", exchangeController.AjaxGetExchanges)
	exchanges.GET("/ajax_getid_exchange", exchangeController.AjaxGetExchangeByID)
	exchanges.POST("/ajax_getid_exchange", exchangeController.AjaxGetExchangeByID)
	exchanges.POST("/ajax_create_exchange", exchangeController.AjaxCreateExchange)
	exchanges.POST("/ajax_edit_exchange", exchangeController.AjaxEditExchange)

	// ============================================
	// ШАГ 9.4: Маршруты для управления аккаунтами бирж (требуют авторизации)
	// ============================================
	exAccounts := r.Group("/exchange_accounts")
	exAccounts.GET("/", exchangeAccountController.List)
	exAccounts.POST("/ajax_get_accounts", exchangeAccountController.AjaxGetAccounts)
	exAccounts.GET("/ajax_getid_accounts", exchangeAccountController.AjaxGetAccountByID)
	exAccounts.POST("/ajax_getid_accounts", exchangeAccountController.AjaxGetAccountByID)
	exAccounts.POST("/ajax_create_account", exchangeAccountController.AjaxCreateAccount)
	exAccounts.POST("/ajax_edit_account", exchangeAccountController.AjaxEditAccount)

	// ============================================
	// ШАГ 10: Настройка статических файлов и шаблонов
	// ============================================
	// Статические файлы (CSS, JS, изображения) будут доступны по пути /assets/*
	// Например: /assets/stylesheets/theme.css
	if !(cfg.Proxy.Enabled && cfg.Proxy.StaticViaNginx) {
		r.Static("/assets", "web/static")
	}

	// ============================================
	// ШАГ 10.0.5: Favicon (браузеры запрашивают автоматически)
	// ============================================
	// Браузеры автоматически запрашивают /favicon.ico при загрузке страницы
	// Отдаём файл из web/static/images/favicon.ico
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("web/static/images/favicon.ico")
	})

	// Загружаем HTML шаблоны
	// Загружаем ВСЕ HTML файлы из root и всех подпапок в один набор шаблонов
	// Это гарантирует что всё загружается один раз и шаблоны не перезаписываются
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse root templates")
	}
	tmpl, err = tmpl.ParseGlob("web/templates/*/*.html")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse subdirectory templates")
	}
	r.SetHTMLTemplate(tmpl)
	logger.Info().Msg("Loaded all HTML templates")

	// ============================================
	// ШАГ 10.1: Обработчик 404 (страница не найдена)
	// ============================================
	// NoRoute обрабатывает все запросы, которые не соответствуют ни одному маршруту
	// Возвращает страницу 404 в едином стиле со страницей авторизации
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	// ============================================
	// ШАГ 11: Запуск HTTP сервера
	// ============================================
	// Формируем адрес для прослушивания по порту
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// Логируем информацию о запуске сервера
	logger.Info().
		Str("address", addr).
		Str("gin_mode", gin.Mode()).
		Str("log_level", cfg.Logging.Level).
		Bool("tls_enabled", cfg.Server.TLS.Enabled).
		Bool("proxy_enabled", cfg.Proxy.Enabled).
		Bool("proxy_trust_forward_headers", cfg.Proxy.TrustForwardHeaders).
		Int("proxy_trusted_hops", cfg.Proxy.TrustedHops).
		Bool("proxy_static_via_nginx", cfg.Proxy.StaticViaNginx).
		Msg("Starting HTTP server")

	// Запускаем сервер и начинаем обрабатывать HTTP запросы
	// r.Run() блокирует выполнение программы - сервер работает до остановки
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       cfg.Server.Timeouts.Read,
		WriteTimeout:      cfg.Server.Timeouts.Write,
		IdleTimeout:       cfg.Server.Timeouts.Idle,
		ReadHeaderTimeout: cfg.Server.Timeouts.ReadHeader,
		MaxHeaderBytes:    cfg.Server.Limits.MaxHeaderBytes,
	}

	if cfg.Server.HTTP2 != nil {
		parsed, err := cfg.Server.HTTP2.Parse()
		if err != nil {
			logger.Fatal().Err(err).Msg("Invalid server.http2 config")
		}

		h2Config := &http2.Server{
			MaxConcurrentStreams:         parsed.MaxConcurrentStreams,
			MaxReadFrameSize:             parsed.MaxFrameSize,
			IdleTimeout:                  time.Duration(parsed.IdleTimeoutSeconds) * time.Second,
			MaxUploadBufferPerConnection: parsed.MaxUploadBufferPerConn,
			MaxUploadBufferPerStream:     parsed.MaxUploadBufferPerStream,
		}
		if err := http2.ConfigureServer(srv, h2Config); err != nil {
			logger.Fatal().Err(err).Msg("Failed to configure HTTP/2 server")
		}
	}

	go func() {
		var serveErr error
		if cfg.Server.TLS.Enabled {
			serveErr = srv.ListenAndServeTLS(cfg.Server.TLS.CertPath, cfg.Server.TLS.KeyPath)
		} else {
			serveErr = srv.ListenAndServe()
		}

		if serveErr != nil && serveErr != http.ErrServerClosed {
			logger.Fatal().
				Err(serveErr).
				Str("address", addr).
				Msg("Failed to start HTTP server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeouts.ShutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("HTTP server graceful shutdown failed")
	} else {
		logger.Info().Msg("HTTP server stopped gracefully")
	}

	// Сюда программа не дойдёт, пока сервер работает
	// Для остановки нужно нажать Ctrl+C или отправить сигнал SIGTERM
}
