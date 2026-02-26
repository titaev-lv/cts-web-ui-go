// Package middleware предоставляет middleware для безопасности приложения.
// Этот файл содержит middleware для защиты от XSS, CSRF и установки security headers.
package middleware

import (
	"ctweb/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"            // CSRF защита
	"github.com/microcosm-cc/bluemonday" // XSS защита (санитизация HTML)
)

var (
	// p - глобальный экземпляр bluemonday для санитизации HTML
	// StrictPolicy() удаляет все HTML теги, оставляя только текст
	// Это защищает от XSS атак при выводе пользовательского контента
	p = bluemonday.StrictPolicy()
)

// SecurityHeadersMiddleware устанавливает HTTP заголовки безопасности.
//
// Что делает:
//
//	Устанавливает стандартные security headers для защиты от различных атак:
//	- XSS Protection
//	- Clickjacking
//	- MIME-type sniffing
//	- и другие
//
// Когда использовать:
//
//	Применяется ко всем HTTP запросам для защиты браузера пользователя.
//
// Пример использования:
//
//	r := gin.Default()
//	r.Use(middleware.SecurityHeadersMiddleware())
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ============================================
		// X-Content-Type-Options: nosniff
		// ============================================
		// Запрещает браузеру "угадывать" тип контента.
		// Без этого заголовка браузер может интерпретировать текстовый файл как HTML,
		// что может привести к XSS атакам.
		c.Header("X-Content-Type-Options", "nosniff")

		// ============================================
		// X-Frame-Options: DENY
		// ============================================
		// Запрещает встраивание страницы в iframe на других сайтах.
		// Защищает от clickjacking атак, когда злоумышленник встраивает
		// вашу страницу в невидимый iframe и пытается заставить пользователя
		// кликнуть по невидимым элементам.
		c.Header("X-Frame-Options", "DENY")

		// ============================================
		// X-XSS-Protection: 1; mode=block
		// ============================================
		// Включает встроенную защиту от XSS в старых браузерах.
		// В современных браузерах (Chrome, Firefox) эта защита встроена,
		// но для совместимости со старыми браузерами стоит оставить.
		c.Header("X-XSS-Protection", "1; mode=block")

		// ============================================
		// Referrer-Policy: strict-origin-when-cross-origin
		// ============================================
		// Контролирует, какую информацию о referrer отправлять.
		// "strict-origin-when-cross-origin" означает:
		//   - При переходе на тот же домен: отправлять полный URL
		//   - При переходе на другой домен: отправлять только origin (домен)
		// Это защищает от утечки информации о внутренней структуре сайта.
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// ============================================
		// Content-Security-Policy (CSP)
		// ============================================
		// Один из самых важных заголовков безопасности.
		// Контролирует, откуда можно загружать ресурсы (JS, CSS, изображения и т.д.).
		//
		// Политика:
		//   - default-src 'self': по умолчанию загружать только с того же домена
		//   - script-src 'self' 'unsafe-inline': разрешить скрипты с того же домена и inline скрипты
		//   - style-src 'self' 'unsafe-inline': разрешить стили с того же домена и inline стили
		//   - img-src 'self' data: https:: разрешить изображения с того же домена, data: и https:
		//   - font-src 'self': разрешить шрифты только с того же домена
		//
		// ВАЖНО: 'unsafe-inline' нужен для работы некоторых библиотек (jQuery, Bootstrap),
		// но в идеале лучше использовать nonce или hash для inline скриптов.
		cfg := config.Get()
		if !cfg.IsDebug() {
			// В продакшн используем более строгую политику
			c.Header("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline'; "+
					"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
					"img-src 'self' data: https:; "+
					"font-src 'self' data: https://fonts.gstatic.com; "+
					"connect-src 'self'; "+
					"frame-ancestors 'none'")
		} else {
			// В режиме разработки более мягкая политика для удобства отладки
			c.Header("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
					"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
					"img-src 'self' data: https:; "+
					"font-src 'self' data: https://fonts.gstatic.com; "+
					"connect-src 'self'")
		}

		// ============================================
		// Permissions-Policy (бывший Feature-Policy)
		// ============================================
		// Контролирует, какие браузерные API можно использовать.
		// Отключаем ненужные функции для повышения безопасности.
		c.Header("Permissions-Policy",
			"geolocation=(), "+
				"microphone=(), "+
				"camera=(), "+
				"payment=(), "+
				"usb=()")

		// Продолжаем обработку запроса
		c.Next()
	}
}

// XSSSanitizeMiddleware санитизирует пользовательский ввод для защиты от XSS.
//
// Что такое XSS (Cross-Site Scripting):
//
//	XSS - это атака, когда злоумышленник внедряет вредоносный JavaScript код
//	в страницу, который выполняется в браузере жертвы.
//
// Пример XSS атаки:
//
//	Пользователь вводит: <script>alert('XSS')</script>
//	Если этот текст выводится без санитизации, скрипт выполнится!
//
// Что делает middleware:
//
//	Удаляет все HTML теги из пользовательского ввода, оставляя только текст.
//	Это защищает от XSS, но также удаляет любое форматирование.
//
// Когда использовать:
//
//	Применяется к запросам, которые принимают пользовательский ввод
//	(формы, API endpoints).
//
// ВАЖНО:
//   - Санитизация применяется только к данным, которые будут выводиться в HTML
//   - Для данных, которые идут в БД, используйте prepared statements (защита от SQL injection)
//   - Для JSON ответов санитизация не нужна (JSON автоматически экранирует)
//
// Пример использования:
//
//	r.POST("/api/users", middleware.XSSSanitizeMiddleware(), createUserHandler)
func XSSSanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Обрабатываем только POST/PUT/PATCH запросы (где может быть пользовательский ввод)
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// Получаем Content-Type
			contentType := c.GetHeader("Content-Type")

			// Санитизируем только form data (application/x-www-form-urlencoded или multipart/form-data)
			if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
				strings.Contains(contentType, "multipart/form-data") {
				// Парсим form data
				if err := c.Request.ParseForm(); err == nil {
					// Санитизируем все значения в форме
					for key, values := range c.Request.PostForm {
						sanitized := make([]string, len(values))
						for i, value := range values {
							// bluemonday.StrictPolicy() удаляет все HTML теги
							sanitized[i] = p.Sanitize(value)
						}
						c.Request.PostForm[key] = sanitized
					}
				}
			}
		}

		c.Next()
	}
}

// SanitizeString санитизирует строку, удаляя все HTML теги.
//
// Используется для санитизации пользовательского ввода перед выводом в HTML.
//
// Параметры:
//   - input: строка, которую нужно санитизировать
//
// Возвращает:
//   - string: санитизированная строка без HTML тегов
//
// Пример использования:
//
//	userInput := "<script>alert('XSS')</script>Hello"
//	safe := middleware.SanitizeString(userInput)
//	// Результат: "Hello" (скрипт удалён)
func SanitizeString(input string) string {
	return p.Sanitize(input)
}

// CSRFMiddleware создаёт CSRF middleware для защиты от CSRF атак.
//
// Что такое CSRF (Cross-Site Request Forgery):
//
//	CSRF - это атака, когда злоумышленник заставляет жертву выполнить
//	нежелательное действие на сайте, где жертва авторизована.
//
// Пример CSRF атаки:
//  1. Пользователь авторизован на сайте bank.com
//  2. Злоумышленник отправляет пользователю ссылку на evil.com
//  3. На evil.com есть форма, которая отправляет POST запрос на bank.com/transfer
//  4. Браузер автоматически отправляет cookies с bank.com, включая сессию
//  5. Запрос выполняется от имени пользователя!
//
// Как защищает CSRF middleware:
//   - Генерирует уникальный токен для каждой сессии
//   - Требует этот токен во всех POST/PUT/DELETE запросах
//   - Проверяет, что токен совпадает с токеном в сессии
//
// Параметры:
//   - secretKey: секретный ключ для генерации токенов (из конфигурации)
//   - cookieName: имя cookie для хранения токена
//   - secure: использовать только HTTPS (true в продакшн)
//
// Возвращает:
//   - gin.HandlerFunc: CSRF middleware
//
// Пример использования:
//
//	csrfMiddleware := middleware.CSRFMiddleware(
//	    config.Get().Security.CSRFSecret,
//	    "csrf_token",
//	    config.Get().Security.SessionCookieSecure,
//	)
//	r.Use(csrfMiddleware)
//
// В HTML формах нужно добавить токен:
//
//	<form method="POST">
//	    <input type="hidden" name="gorilla.csrf.Token" value="{{.csrf_token}}">
//	    ...
//	</form>
func CSRFMiddleware(secretKey, cookieName string, secure bool, headerName string, trustedOrigins []string) gin.HandlerFunc {
	if headerName == "" {
		headerName = "X-CSRF-Token"
	}

	options := []csrf.Option{
		csrf.Secure(secure),                  // Использовать только HTTPS (если secure = true)
		csrf.HttpOnly(true),                  // Cookie недоступен через JavaScript (защита от XSS)
		csrf.SameSite(csrf.SameSiteLaxMode),  // Политика SameSite
		csrf.CookieName(cookieName),          // Имя cookie
		csrf.FieldName("gorilla.csrf.Token"), // Имя поля в форме
		csrf.RequestHeader(headerName),       // Имя заголовка для AJAX запросов
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Обработчик ошибки CSRF
			// Если токен неверен или отсутствует, возвращаем 403 Forbidden
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("CSRF token validation failed"))
		})),
	}
	if len(trustedOrigins) > 0 {
		options = append(options, csrf.TrustedOrigins(trustedOrigins))
	}

	// Создаём CSRF middleware с настройками
	csrfProtect := csrf.Protect(
		[]byte(secretKey), // Секретный ключ для генерации токенов
		options...,
	)

	// Обёртка для Gin
	return func(c *gin.Context) {
		// CSRF middleware работает с http.Handler, поэтому используем обёртку
		csrfProtect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Сохраняем токен в контексте Gin для использования в шаблонах
			token := csrf.Token(r)
			c.Set("csrf_token", token)
			c.Set(headerName, token) // Также в заголовке для AJAX запросов

			// Продолжаем обработку
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

// GetCSRFToken возвращает CSRF токен из контекста Gin.
//
// Используется в шаблонах для добавления токена в формы.
//
// Параметры:
//   - c: контекст Gin
//
// Возвращает:
//   - string: CSRF токен или пустую строку, если токен не установлен
//
// Пример использования в шаблоне:
//
//	<form method="POST">
//	    <input type="hidden" name="gorilla.csrf.Token" value="{{.csrf_token}}">
//	    ...
//	</form>
func GetCSRFToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		if tokenStr, ok := token.(string); ok {
			return tokenStr
		}
	}
	return ""
}
