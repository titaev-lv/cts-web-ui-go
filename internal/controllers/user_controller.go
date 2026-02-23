// Package controllers содержит HTTP handlers (контроллеры) для обработки запросов.
// Контроллеры получают запросы, вызывают сервисы и возвращают ответы.
package controllers

import (
	"ctweb/internal/errors"       // Централизованная обработка ошибок
	"ctweb/internal/logger"       // Система логирования
	"ctweb/internal/middleware"   // Middleware для получения пользователя
	"ctweb/internal/models"       // Модели данных
	"ctweb/internal/repositories" // Репозитории для работы с БД
	"ctweb/internal/services"
	"ctweb/internal/session" // Управление сессиями
	"ctweb/internal/utils"   // Утилиты (DataTables парсинг, валидация)
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController - контроллер для работы с пользователями.
// Содержит методы для обработки HTTP запросов, связанных с пользователями.
type UserController struct{}

// NewUserController создаёт новый экземпляр UserController.
//
// Возвращает:
//   - *UserController: новый контроллер
func NewUserController() *UserController {
	return &UserController{}
}

// Home обрабатывает запрос на главную страницу.
//
// Параметры:
//   - c: контекст Gin (содержит запрос и ответ)
//
// Что делает:
//
//	Рендерит HTML шаблон главной страницы с базовым layout.
//	Передаёт информацию о пользователе (если авторизован) в шаблон.
//
// Шаблоны:
//   - index.html - контент главной страницы
//
// Home отображает главную страницу системы.
// Получает текущего пользователя из контекста (установлен в AuthMiddleware)
func (u *UserController) Home(c *gin.Context) {
	// Получаем текущего пользователя из контекста
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		// Не должно произойти, т.к. middleware проверяет аутентификацию
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Рендерим template с данными пользователя
	c.HTML(http.StatusOK, "home", gin.H{
		"Title": "Home",
		"User":  user,
	})
}

// ShowLoginPage отображает страницу входа.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//
//	Рендерит HTML шаблон login.html.
//
// Шаблон должен находиться в web/templates/login.html
func (u *UserController) ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// Login обрабатывает запрос на вход пользователя.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Парсит JSON с логином и паролем
//  2. Проверяет креды пользователя
//  3. Генерирует JWT токен
//  4. Возвращает токен в формате, совместимом с PHP
//
// Формат запроса:
//
//	POST /auth/login
//	{
//	  "username": "user123",
//	  "password": "password123"
//	}
//
// Формат ответа (успех):
//
//	{
//	  "error": false,
//	  "success": true,
//	  "data": {
//	    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//	  }
//	}
//
// Формат ответа (ошибка):
//
//	{
//	  "error": "Invalid credentials",
//	  "success": false
//	}
func (u *UserController) Login(c *gin.Context) {
	// ============================================
	// ШАГ 1: Парсинг form data (как в PHP)
	// ============================================
	// PHP форма отправляет данные как application/x-www-form-urlencoded
	// Поля: username, pwd, rememberme
	username := c.PostForm("username")           // Логин пользователя
	password := c.PostForm("pwd")                // Пароль пользователя (в PHP это 'pwd', не 'password')
	remember := c.PostForm("rememberme") == "on" // Remember Me (checkbox)

	// ============================================
	// ШАГ 2: Валидация входных данных
	// ============================================

	// Проверяем, что логин и пароль не пустые
	if username == "" {
		errors.HandleError(c, errors.ValidationError("Username is required", nil))
		return
	}

	if password == "" {
		errors.HandleError(c, errors.ValidationError("Password is required", nil))
		return
	}

	// ============================================
	// ШАГ 3: Аутентификация через AuthService
	// ============================================
	// Используем AuthService для входа (он использует bcrypt и загружает группы)
	authService := services.NewAuthService()
	authServiceStart := time.Now()
	result := authService.Login(username, password, remember)
	middleware.AddLatencyPart(c, "auth_service_call_ms", time.Since(authServiceStart))
	for key, value := range result.Timings {
		middleware.AddLatencyPartMS(c, key, value)
	}

	// Проверяем результат
	if result.Error != nil {
		// Дополнительное логирование в контроллере с IP адресом и User-Agent
		// Это помогает отслеживать подозрительную активность
		// ВАЖНО: НЕ логируем пароль!
		logger.Warn().
			Str("login", username).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.GetHeader("User-Agent")).
			Str("error", result.Error.Error()).
			Str("event", "login_failed").
			Msg("Login failed")

		// Формат ответа как в PHP: {"error": "текст ошибки", "success": false}
		// Получаем текст ошибки из AppError
		errorMsg := result.Error.Error()
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Установка сессии и cookies
	// ============================================
	// Сохраняем данные пользователя в сессию (как в PHP: $_SESSION['ct_user'])
	sm := session.GetSessionManager()
	sessionStart := time.Now()
	if err := sm.SetUser(c.Request, c.Writer, result.User); err != nil {
		middleware.AddLatencyPart(c, "session_write_ms", time.Since(sessionStart))
		logger.Error().
			Err(err).
			Int("user_id", result.User.ID).
			Msg("Failed to set session")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed to create session",
			"success": false,
		})
		return
	}

	// Если выбрано "Remember Me", устанавливаем cookies с токеном
	if remember && result.Token != "" {
		rememberStart := time.Now()
		sm.SetRememberMeCookies(c.Request, c.Writer, username, result.Token)
		middleware.AddLatencyPart(c, "remember_cookie_ms", time.Since(rememberStart))
		logger.Debug().
			Str("login", username).
			Msg("Remember Me cookies set")
	}

	middleware.AddLatencyPart(c, "session_write_ms", time.Since(sessionStart))

	// ============================================
	// ШАГ 5: Логирование успешного входа
	// ============================================
	// Логируем успешный вход с метаданными для аудита
	// ВАЖНО: НЕ логируем пароль!
	logger.Info().
		Str("login", username).
		Int("user_id", result.User.ID).
		Str("client_ip", c.ClientIP()).
		Str("user_agent", c.GetHeader("User-Agent")).
		Bool("remember_me", remember).
		Str("event", "login_success").
		Msg("User logged in successfully")

	// ============================================
	// ШАГ 6: Отправка успешного ответа
	// ============================================
	// Формат ответа как в PHP: {"error": false, "success": true}
	// В PHP не возвращается токен в ответе, он сохраняется в сессии/куках
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}

// Logout обрабатывает запрос на выход из системы.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Получает пользователя из контекста (установлен в AuthMiddleware)
//  2. Вызывает AuthService.Logout() для удаления токена из БД
//  3. Вызывает SessionManager.ClearUser() для очистки сессии и cookies
//  4. Редиректит на страницу входа
//
// Использование:
//
//	GET /auth/logout - выход из системы
//
// Примечание:
//
//	После выхода пользователь будет перенаправлен на /login
func (u *UserController) Logout(c *gin.Context) {
	// ============================================
	// ШАГ 1: Получение пользователя из контекста
	// ============================================
	// Пользователь должен быть установлен в AuthMiddleware
	// Если пользователя нет, всё равно выполняем выход (на случай частичной сессии)
	user, exists := middleware.GetUserFromContext(c)

	var userID int
	if exists && user != nil {
		userID = user.ID
		logger.Debug().
			Int("user_id", userID).
			Str("login", user.Login).
			Msg("Logout requested")
	} else {
		logger.Debug().
			Msg("Logout requested but user not found in context")
	}

	// ============================================
	// ШАГ 2: Удаление токена из БД
	// ============================================
	// Если пользователь найден, удаляем токен "Remember Me" из БД
	// Это делает невозможным восстановление сессии из cookies
	if exists && user != nil {
		authService := services.NewAuthService()
		err := authService.Logout(userID)
		if err != nil {
			// Ошибка при удалении токена - логируем, но продолжаем выход
			// Сессия всё равно будет очищена
			logger.Error().
				Err(err).
				Int("user_id", userID).
				Msg("Failed to clear remember me token on logout")
		}
	}

	// ============================================
	// ШАГ 3: Очистка сессии и cookies
	// ============================================
	// Очищаем сессию и удаляем cookies "Remember Me" (Login и CTToken)
	// Это делается всегда, даже если пользователь не найден в контексте
	sm := session.GetSessionManager()
	if err := sm.ClearUser(c.Request, c.Writer); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to clear session on logout")
		// Продолжаем выход, даже если не удалось очистить сессию
	}

	// ============================================
	// ШАГ 4: Логирование успешного выхода
	// ============================================
	if exists && user != nil {
		logger.Info().
			Int("user_id", userID).
			Str("login", user.Login).
			Str("client_ip", c.ClientIP()).
			Str("event", "logout_success").
			Msg("User logged out successfully")
	}

	// ============================================
	// ШАГ 5: Редирект на страницу входа
	// ============================================
	// Редиректим на страницу входа (как в PHP после logout)
	c.Redirect(http.StatusFound, "/login")
}

// AjaxGetUsers обрабатывает AJAX запрос от DataTables для получения списка пользователей.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Парсит параметры DataTables запроса
//  3. Вызывает репозиторий для получения данных
//  4. Форматирует ответ в формате DataTables (aaData)
//
// Формат запроса:
//
//	POST /users/ajax_get_users
//	Параметры DataTables (draw, start, length, search[value], order[0][column], columns[i][data] и т.д.)
//
// Формат ответа:
//
//	{
//	  "draw": 1,
//	  "recordsTotal": 100,
//	  "recordsFiltered": 50,
//	  "aaData": [
//	    {
//	      "chbx": "",
//	      "DT_RowId": "row_1",
//	      "id": 1,
//	      "login": "admin",
//	      "groups": "Admin",
//	      "active": "Active",
//	      "name": "Doe John",
//	      "email": "admin@example.com",
//	      "create_date": "01-12-2025 10:30:00",
//	      "modify_date": "02-12-2025 15:20:00",
//	      "timestamp": "03-12-2025 09:15:00"
//	    }
//	  ]
//	}
func (u *UserController) AjaxGetUsers(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров DataTables запроса
	// ============================================
	req := utils.ParseDataTablesRequest(c)

	// ============================================
	// ШАГ 3: Получение данных из репозитория
	// ============================================
	userRepo := repositories.NewUserRepository()

	// Конвертируем запрос в формат репозитория
	userReq := utils.ConvertToUserRepositoryRequest(req)

	// Получаем данные
	repoResponse, err := userRepo.FindAllWithPagination(userReq)
	if err != nil {
		logger.Error().
			Err(err).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to get users list for DataTables")

		// Возвращаем пустой ответ в случае ошибки (как в PHP)
		c.JSON(http.StatusOK, gin.H{
			"recordsTotal":    0,
			"recordsFiltered": 0,
			"aaData":          []interface{}{},
		})
		return
	}

	// ============================================
	// ШАГ 4: Форматирование ответа в формат DataTables
	// ============================================
	response := utils.ConvertUserResponseToDataTablesFormat(req.Draw, repoResponse)

	// ============================================
	// ШАГ 5: Возврат ответа
	// ============================================
	c.JSON(http.StatusOK, response)
}

// List отображает страницу управления пользователями.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Рендерит HTML шаблон страницы пользователей
//
// Формат запроса:
//
//	GET /users/
//
// Шаблон:
//   - users/index.html (если будет создан)
func (u *UserController) List(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ============================================
	// ШАГ 2: Загрузка всех групп для выпадающих списков
	// ============================================
	groupRepo := repositories.NewGroupRepository()
	groups, err := groupRepo.FindAll()
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to load groups for users page")
		// Продолжаем с пустым списком групп
		groups = []*models.Group{}
	}

	// ============================================
	// ШАГ 3: Рендеринг страницы
	// ============================================
	c.HTML(http.StatusOK, "users/index.html", gin.H{
		"Title":  "Users Management",
		"User":   user,
		"Groups": groups,
	})
}

// AjaxGetUserById обрабатывает AJAX запрос для получения данных пользователя по ID.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Получает ID пользователя из параметров запроса
//  3. Загружает пользователя из БД вместе с группами
//  4. Форматирует ответ в формате PHP (lowercase поля)
//
// Формат запроса:
//
//	GET /users/ajax_getid_user?id=1
//	или
//	POST /users/ajax_getid_user с параметром id
//
// Формат ответа:
//
//	{
//	  "error": false,
//	  "success": true,
//	  "data": {
//	    "id": "1",
//	    "login": "admin",
//	    "groups": "1,2",
//	    "status": "enable",
//	    "last_name": "Doe",
//	    "name": "John",
//	    "email": "admin@example.com"
//	  }
//	}
func (u *UserController) AjaxGetUserById(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Получение ID пользователя из параметров
	// ============================================
	// Поддерживаем оба формата: GET (query) и POST (form)
	var userIDStr string
	if c.Request.Method == "GET" {
		userIDStr = c.Query("id")
	} else {
		userIDStr = c.PostForm("id")
	}

	if userIDStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   "User ID is required",
			"success": false,
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 3: Получение пользователя из БД
	// ============================================
	userRepo := repositories.NewUserRepository()
	targetUser, err := userRepo.FindByID(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to get user by ID")

		c.JSON(http.StatusOK, gin.H{
			"error":   "DB ERROR: Error select user",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Загрузка групп пользователя
	// ============================================
	groups, err := userRepo.FindGroupsByUserID(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to get user groups")
		groups = []int{} // Продолжаем с пустым списком групп
	}

	// Конвертируем группы в строку через запятую (как в PHP)
	groupsStr := ""
	if len(groups) > 0 {
		groupStrs := make([]string, len(groups))
		for i, gid := range groups {
			groupStrs[i] = strconv.Itoa(gid)
		}
		groupsStr = strings.Join(groupStrs, ",")
	}

	// ============================================
	// ШАГ 5: Форматирование ответа (как в PHP - lowercase поля)
	// ============================================
	status := utils.BoolToStatus(targetUser.Active)

	responseData := gin.H{
		"id":        strconv.Itoa(targetUser.ID),
		"login":     targetUser.Login,
		"groups":    groupsStr,
		"status":    status,
		"last_name": targetUser.LastName,
		"name":      targetUser.Name,
		"email":     targetUser.Email,
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data":    responseData,
	})
}

// AjaxCreateUser обрабатывает AJAX запрос для создания нового пользователя.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Парсит и валидирует все поля формы
//  3. Проверяет уникальность логина и email
//  4. Валидирует пароль
//  5. Создаёт пользователя в БД
//
// Формат запроса:
//
//	POST /users/ajax_create_user
//	Параметры:
//	  - create_user_login: логин
//	  - create_user_password: пароль
//	  - create_user_password_confirm: подтверждение пароля
//	  - create_user_email: email
//	  - create_user_groups: группы через запятую (например, "1,2")
//	  - create_user_status: статус ("enable" или "disable")
//	  - create_user_name: имя
//	  - create_user_last_name: фамилия
//
// Формат ответа:
//
//	{
//	  "error": false,
//	  "success": true
//	}
//	или
//	{
//	  "error": "Error message",
//	  "success": false
//	}
func (u *UserController) AjaxCreateUser(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров формы
	// ============================================
	login := strings.TrimSpace(c.PostForm("create_user_login"))
	password := c.PostForm("create_user_password")
	passwordConfirm := c.PostForm("create_user_password_confirm")
	email := strings.TrimSpace(c.PostForm("create_user_email"))
	groupsStr := c.PostForm("create_user_groups")
	status := c.PostForm("create_user_status")
	name := strings.TrimSpace(c.PostForm("create_user_name"))
	lastName := strings.TrimSpace(c.PostForm("create_user_last_name"))

	// ============================================
	// ШАГ 3: Валидация обязательных полей
	// ============================================
	var errorMsg string

	if login == "" {
		errorMsg = "Field \"Login\" is empty"
	} else if password == "" {
		errorMsg = "Field \"Password\" is empty"
	} else if passwordConfirm == "" {
		errorMsg = "Field \"Password Confirm\" is empty"
	} else if email == "" {
		errorMsg = "Field \"Email\" is empty"
	} else if groupsStr == "" {
		errorMsg = "Field \"Groups\" is empty"
	} else if status == "" {
		errorMsg = "Field \"Status\" is empty"
	} else if name == "" {
		errorMsg = "Field \"Name\" is empty"
	} else if lastName == "" {
		errorMsg = "Field \"Last Name\" is empty"
	}

	if errorMsg != "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Валидация пароля
	// ============================================
	if err := utils.PasswordValidateWithConfirm(password, passwordConfirm); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 5: Валидация статуса
	// ============================================
	if err := utils.ValidateStatus(status); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6: Валидация email
	// ============================================
	if err := utils.ValidateEmail(email); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6.1: Создание репозитория для проверки уникальности
	// ============================================
	userRepo := repositories.NewUserRepository()

	// ============================================
	// ШАГ 6.2: Проверка уникальности email
	// ============================================
	existsEmail, err := userRepo.ExistsByEmail(email)
	if err != nil {
		logger.Error().
			Err(err).
			Str("email", email).
			Msg("Failed to check email uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsEmail {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Email already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7: Проверка уникальности логина
	// ============================================
	existsLogin, err := userRepo.ExistsByLogin(login)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Msg("Failed to check login uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsLogin {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Login already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 8: Парсинг и валидация групп
	// ============================================
	groupIDs, err := utils.ParseGroupIDs(groupsStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// Проверяем, что все группы существуют
	if err := utils.ValidateGroupIDs(groupIDs); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed groups define",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 9: Хеширование пароля
	// ============================================
	hashedPassword, err := utils.PasswordHash(password)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to hash password")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed to hash password",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 10: Создание пользователя
	// ============================================
	newUser := &models.User{
		Login:    login,
		Password: hashedPassword,
		Email:    email,
		Active:   utils.StatusToBool(status),
		Name:     name,
		LastName: lastName,
		Timezone: "UTC", // По умолчанию UTC
	}

	createdUserID, err := userRepo.Create(newUser, groupIDs, user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to create user")

		c.JSON(http.StatusOK, gin.H{
			"error":   "Error create user",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 11: Логирование успешного создания
	// ============================================
	logger.Info().
		Int("created_user_id", createdUserID).
		Str("login", login).
		Int("creator_user_id", user.ID).
		Str("client_ip", c.ClientIP()).
		Msg("User created successfully")

	// ============================================
	// ШАГ 12: Возврат успешного ответа
	// ============================================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}

// AjaxEditUser обрабатывает AJAX запрос для редактирования пользователя.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//  1. Проверяет, что пользователь - администратор
//  2. Парсит и валидирует все поля формы
//  3. Проверяет уникальность логина и email (исключая текущего пользователя)
//  4. Валидирует пароль (если указан)
//  5. Обновляет пользователя в БД
//
// Формат запроса:
//
//	POST /users/ajax_edit_user
//	Параметры:
//	  - edit_user_id: ID пользователя
//	  - edit_user_login: логин
//	  - edit_user_password: пароль (опционально, если пустой - не обновляется)
//	  - edit_user_password_confirm: подтверждение пароля (если пароль указан)
//	  - edit_user_email: email
//	  - edit_user_groups: группы через запятую (например, "1,2")
//	  - edit_user_status: статус ("enable" или "disable")
//	  - edit_user_name: имя
//	  - edit_user_last_name: фамилия
//
// Формат ответа:
//
//	{
//	  "error": false,
//	  "success": true
//	}
//	или
//	{
//	  "error": "Error message",
//	  "success": false
//	}
func (u *UserController) AjaxEditUser(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров формы
	// ============================================
	userIDStr := c.PostForm("edit_user_id")
	login := strings.TrimSpace(c.PostForm("edit_user_login"))
	password := c.PostForm("edit_user_password")
	passwordConfirm := c.PostForm("edit_user_password_confirm")
	email := strings.TrimSpace(c.PostForm("edit_user_email"))
	groupsStr := c.PostForm("edit_user_groups")
	status := c.PostForm("edit_user_status")
	name := strings.TrimSpace(c.PostForm("edit_user_name"))
	lastName := strings.TrimSpace(c.PostForm("edit_user_last_name"))

	// ============================================
	// ШАГ 3: Валидация обязательных полей
	// ============================================
	var errorMsg string

	if userIDStr == "" {
		errorMsg = "Field \"User ID\" is empty"
	} else if login == "" {
		errorMsg = "Field \"Login\" is empty"
	} else if email == "" {
		errorMsg = "Field \"Email\" is empty"
	} else if groupsStr == "" {
		errorMsg = "Field \"Groups\" is empty"
	} else if status == "" {
		errorMsg = "Field \"Status\" is empty"
	} else if name == "" {
		errorMsg = "Field \"Name\" is empty"
	} else if lastName == "" {
		errorMsg = "Field \"Last Name\" is empty"
	}

	// Если пароль указан, проверяем и подтверждение
	if password != "" && passwordConfirm == "" {
		errorMsg = "Field \"Password Confirm\" is empty"
	}

	if errorMsg != "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Invalid user ID",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Проверка уникальности логина (исключая текущего пользователя)
	// ============================================
	userRepo := repositories.NewUserRepository()
	existsLogin, err := userRepo.ExistsByLoginExcludingID(login, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("login", login).
			Int("user_id", userID).
			Msg("Failed to check login uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsLogin {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Login already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 5: Валидация пароля (если указан)
	// ============================================
	updatePassword := password != ""
	if updatePassword {
		if err := utils.PasswordValidateWithConfirm(password, passwordConfirm); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error":   err.Error(),
				"success": false,
			})
			return
		}
	}

	// ============================================
	// ШАГ 6: Валидация статуса
	// ============================================
	if err := utils.ValidateStatus(status); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7: Валидация email
	// ============================================
	if err := utils.ValidateEmail(email); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7.1: Проверка уникальности email (исключая текущего пользователя)
	// ============================================
	existsEmail, err := userRepo.ExistsByEmailExcludingID(email, userID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("email", email).
			Int("user_id", userID).
			Msg("Failed to check email uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsEmail {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Email already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 8: Парсинг и валидация групп
	// ============================================
	groupIDs, err := utils.ParseGroupIDs(groupsStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// Проверяем, что все группы существуют
	if err := utils.ValidateGroupIDs(groupIDs); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Failed groups define",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 9: Получение существующего пользователя
	// ============================================
	existingUser, err := userRepo.FindByID(userID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Msg("Failed to find user for update")
		c.JSON(http.StatusOK, gin.H{
			"error":   "User not found",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 10: Хеширование пароля (если указан)
	// ============================================
	if updatePassword {
		hashedPassword, err := utils.PasswordHash(password)
		if err != nil {
			logger.Error().
				Err(err).
				Msg("Failed to hash password")
			c.JSON(http.StatusOK, gin.H{
				"error":   "Failed to hash password",
				"success": false,
			})
			return
		}
		existingUser.Password = hashedPassword
	}

	// ============================================
	// ШАГ 11: Обновление данных пользователя
	// ============================================
	existingUser.Login = login
	existingUser.Email = email
	existingUser.Active = utils.StatusToBool(status)
	existingUser.Name = name
	existingUser.LastName = lastName

	err = userRepo.Update(existingUser, groupIDs, updatePassword, user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", userID).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to update user")

		c.JSON(http.StatusOK, gin.H{
			"error":   "Error edit user",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 12: Логирование успешного обновления
	// ============================================
	logger.Info().
		Int("updated_user_id", userID).
		Str("login", login).
		Int("editor_user_id", user.ID).
		Str("client_ip", c.ClientIP()).
		Msg("User updated successfully")

	// ============================================
	// ШАГ 13: Возврат успешного ответа
	// ============================================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}
