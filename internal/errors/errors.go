// Package errors предоставляет централизованную обработку ошибок приложения.
// Все ошибки обрабатываются единообразно с логированием и правильными HTTP статусами.
package errors

import (
	"ctweb/internal/logger"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AppError - структура для представления ошибки приложения.
//
// Содержит всю информацию об ошибке:
//   - Code: HTTP статус код
//   - Message: сообщение для пользователя
//   - InternalError: внутренняя ошибка (для логирования, не показывается пользователю)
//   - Details: дополнительные детали (опционально)
type AppError struct {
	Code         int                    `json:"code"`          // HTTP статус код (400, 404, 500 и т.д.)
	Message      string                 `json:"message"`       // Сообщение для пользователя
	InternalError error                 `json:"-"`            // Внутренняя ошибка (не отправляется клиенту)
	Details      map[string]interface{} `json:"details,omitempty"` // Дополнительные детали (опционально)
}

// Error реализует интерфейс error.
// Возвращает сообщение об ошибке.
func (e *AppError) Error() string {
	if e.InternalError != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.InternalError)
	}
	return e.Message
}

// ============================================
// ПРЕДОПРЕДЕЛЁННЫЕ ТИПЫ ОШИБОК
// ============================================

// ValidationError создаёт ошибку валидации (400 Bad Request).
//
// Используется, когда пользовательский ввод не прошёл валидацию.
//
// Параметры:
//   - message: сообщение об ошибке валидации
//   - details: дополнительные детали (например, какие поля невалидны)
//
// Возвращает:
//   - *AppError: ошибка валидации
//
// Пример использования:
//
//	if username == "" {
//	    return nil, errors.ValidationError("Username is required", nil)
//	}
//
//	if len(password) < 10 {
//	    return nil, errors.ValidationError("Password too short", map[string]interface{}{
//	        "field": "password",
//	        "min_length": 10,
//	    })
//	}
func ValidationError(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Details: details,
	}
}

// NotFoundError создаёт ошибку "не найдено" (404 Not Found).
//
// Используется, когда запрашиваемый ресурс не существует.
//
// Параметры:
//   - resource: тип ресурса (например, "user", "group")
//   - id: идентификатор ресурса (опционально)
//
// Возвращает:
//   - *AppError: ошибка "не найдено"
//
// Пример использования:
//
//	user, err := getUserByID(userID)
//	if err == sql.ErrNoRows {
//	    return nil, errors.NotFoundError("user", userID)
//	}
func NotFoundError(resource string, id interface{}) *AppError {
	message := fmt.Sprintf("%s not found", strings.Title(resource))
	if id != nil {
		message = fmt.Sprintf("%s with ID %v not found", strings.Title(resource), id)
	}
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
		Details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

// UnauthorizedError создаёт ошибку авторизации (401 Unauthorized).
//
// Используется, когда пользователь не авторизован или неверные креды.
//
// Параметры:
//   - message: сообщение об ошибке (по умолчанию "Unauthorized")
//
// Возвращает:
//   - *AppError: ошибка авторизации
//
// Пример использования:
//
//	if !user.IsActive {
//	    return nil, errors.UnauthorizedError("User is blocked")
//	}
//
//	if !checkPassword(password, user.Password) {
//	    return nil, errors.UnauthorizedError("Invalid credentials")
//	}
func UnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// ForbiddenError создаёт ошибку доступа (403 Forbidden).
//
// Используется, когда пользователь авторизован, но не имеет прав доступа.
//
// Параметры:
//   - message: сообщение об ошибке (по умолчанию "Forbidden")
//
// Возвращает:
//   - *AppError: ошибка доступа
//
// Пример использования:
//
//	if !user.IsAdmin() {
//	    return nil, errors.ForbiddenError("Admin access required")
//	}
func ForbiddenError(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// ConflictError создаёт ошибку конфликта (409 Conflict).
//
// Используется, когда операция конфликтует с текущим состоянием (например, дубликат).
//
// Параметры:
//   - message: сообщение об ошибке
//
// Возвращает:
//   - *AppError: ошибка конфликта
//
// Пример использования:
//
//	if userExists(login) {
//	    return nil, errors.ConflictError("User with this login already exists")
//	}
func ConflictError(message string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Message: message,
	}
}

// InternalError создаёт внутреннюю ошибку сервера (500 Internal Server Error).
//
// Используется для неожиданных ошибок, которые не должны показываться пользователю.
//
// Параметры:
//   - message: сообщение для пользователя (общее, без деталей)
//   - err: внутренняя ошибка (будет залогирована, но не показана пользователю)
//
// Возвращает:
//   - *AppError: внутренняя ошибка
//
// Пример использования:
//
//	result, err := db.Exec("INSERT INTO users ...")
//	if err != nil {
//	    return nil, errors.InternalError("Failed to create user", err)
//	}
func InternalError(message string, err error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:          http.StatusInternalServerError,
		Message:       message,
		InternalError: err,
	}
}

// DatabaseError обрабатывает ошибки базы данных и преобразует их в AppError.
//
// Что делает:
//   - sql.ErrNoRows -> NotFoundError
//   - Остальные ошибки БД -> InternalError
//
// Параметры:
//   - resource: тип ресурса (для NotFoundError)
//   - id: идентификатор ресурса (опционально)
//   - err: ошибка из базы данных
//
// Возвращает:
//   - *AppError: обработанная ошибка
//
// Пример использования:
//
//	user, err := getUserByID(userID)
//	if err != nil {
//	    return nil, errors.DatabaseError("user", userID, err)
//	}
func DatabaseError(resource string, id interface{}, err error) *AppError {
	if err == nil {
		return nil
	}

	// Если запись не найдена, возвращаем NotFoundError
	if errors.Is(err, sql.ErrNoRows) {
		return NotFoundError(resource, id)
	}

	// Остальные ошибки БД - это внутренние ошибки
	return InternalError(fmt.Sprintf("Database error while accessing %s", resource), err)
}

// ============================================
// ОБРАБОТКА ОШИБОК В GIN
// ============================================

// HandleError обрабатывает ошибку и отправляет ответ клиенту.
//
// Что делает:
//   1. Логирует ошибку (если это внутренняя ошибка)
//   2. Формирует JSON ответ с ошибкой
//   3. Устанавливает правильный HTTP статус код
//   4. Отправляет ответ клиенту
//
// Параметры:
//   - c: контекст Gin
//   - err: ошибка для обработки
//
// Пример использования:
//
//	user, err := createUser(data)
//	if err != nil {
//	    errors.HandleError(c, err)
//	    return
//	}
func HandleError(c *gin.Context, err error) {
	// Если ошибка nil, ничего не делаем
	if err == nil {
		return
	}

	var appErr *AppError

	// Проверяем, является ли ошибка AppError
	if errors.As(err, &appErr) {
		// Это наша ошибка, обрабатываем её
		handleAppError(c, appErr)
		return
	}

	// Если это не AppError, создаём InternalError
	appErr = InternalError("An unexpected error occurred", err)
	handleAppError(c, appErr)
}

// handleAppError обрабатывает AppError.
//
// Внутренняя функция, вызывается из HandleError.
func handleAppError(c *gin.Context, err *AppError) {
	// Логируем ошибку
	// Для внутренних ошибок логируем с деталями
	// Для пользовательских ошибок (400, 404) логируем на уровне Warn
	if err.Code >= 500 {
		// Внутренняя ошибка сервера - логируем с полными деталями
		logger.Error().
			Err(err.InternalError).
			Int("status_code", err.Code).
			Str("message", err.Message).
			Interface("details", err.Details).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Msg("Internal server error")
	} else if err.Code >= 400 {
		// Пользовательская ошибка (400, 401, 403, 404) - логируем на уровне Warn
		logger.Warn().
			Int("status_code", err.Code).
			Str("message", err.Message).
			Interface("details", err.Details).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Str("client_ip", c.ClientIP()).
			Msg("Client error")
	}

	// Формируем ответ в формате, совместимом с PHP кодом
	// PHP возвращает: {"error": "...", "success": false}
	response := gin.H{
		"error":   err.Message,
		"success": false,
	}

	// Добавляем детали, если они есть
	if err.Details != nil && len(err.Details) > 0 {
		response["details"] = err.Details
	}

	// Отправляем ответ с правильным статус кодом
	c.JSON(err.Code, response)
	c.Abort() // Прерываем выполнение цепочки middleware
}

// HandleSuccess отправляет успешный ответ в формате, совместимом с PHP.
//
// PHP возвращает: {"error": false, "success": true, "data": ...}
//
// Параметры:
//   - c: контекст Gin
//   - data: данные для отправки (опционально)
//
// Пример использования:
//
//	user, err := createUser(data)
//	if err != nil {
//	    errors.HandleError(c, err)
//	    return
//	}
//	errors.HandleSuccess(c, gin.H{"user_id": user.ID})
func HandleSuccess(c *gin.Context, data interface{}) {
	response := gin.H{
		"error":   false,
		"success": true,
	}

	if data != nil {
		response["data"] = data
	}

	c.JSON(http.StatusOK, response)
}

// ============================================
// УДОБНЫЕ ФУНКЦИИ ДЛЯ РАЗНЫХ СЛУЧАЕВ
// ============================================

// HandleValidationErrors обрабатывает множественные ошибки валидации.
//
// Используется, когда нужно вернуть несколько ошибок валидации одновременно.
//
// Параметры:
//   - c: контекст Gin
//   - errors: карта ошибок (поле -> сообщение)
//
// Пример использования:
//
//	validationErrors := map[string]string{
//	    "username": "Username is required",
//	    "email":    "Invalid email format",
//	}
//	if len(validationErrors) > 0 {
//	    errors.HandleValidationErrors(c, validationErrors)
//	    return
//	}
func HandleValidationErrors(c *gin.Context, validationErrors map[string]string) {
	details := make(map[string]interface{})
	for field, message := range validationErrors {
		details[field] = message
	}

	err := ValidationError("Validation failed", details)
	handleAppError(c, err)
}

// HandleDatabaseError - удобная функция для обработки ошибок БД.
//
// Автоматически определяет тип ошибки и создаёт соответствующий AppError.
//
// Параметры:
//   - c: контекст Gin
//   - resource: тип ресурса
//   - id: идентификатор ресурса
//   - err: ошибка из БД
//
// Пример использования:
//
//	user, err := getUserByID(userID)
//	if err != nil {
//	    errors.HandleDatabaseError(c, "user", userID, err)
//	    return
//	}
func HandleDatabaseError(c *gin.Context, resource string, id interface{}, err error) {
	if err == nil {
		return
	}

	appErr := DatabaseError(resource, id, err)
	handleAppError(c, appErr)
}

