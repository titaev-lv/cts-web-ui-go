// Package middleware - Recovery middleware для обработки паник (panic).
// Этот файл содержит middleware для перехвата паник и предотвращения краша приложения.
package middleware

import (
	"ctweb/internal/errors" // Централизованная обработка ошибок
	"ctweb/internal/logger" // Система логирования
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware создаёт middleware для обработки паник.
//
// Что такое panic в Go:
//
//	Panic - это критическая ошибка, которая останавливает выполнение программы.
//	Без обработки panic приложение крашится.
//
// Что делает middleware:
//  1. Перехватывает panic с помощью recover()
//  2. Логирует панику с полным stack trace
//  3. Возвращает пользователю корректный ответ (500 Internal Server Error)
//  4. Предотвращает краш приложения
//
// Когда использовать:
//
//	Должен быть первым middleware в цепочке, чтобы перехватывать паники
//	из всех последующих middleware и handlers.
//
// Пример использования:
//
//	r := gin.New() // gin.New() без встроенного Recovery
//	r.Use(middleware.RecoveryMiddleware())
//
// Или заменить встроенный Recovery:
//
//	r := gin.Default() // gin.Default() уже имеет Recovery
//	// Заменяем встроенный Recovery на наш
//	r.Use(middleware.RecoveryMiddleware())
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ============================================
		// ОТЛОЖЕННАЯ ФУНКЦИЯ ДЛЯ ПЕРЕХВАТА PANIC
		// ============================================
		// defer означает, что функция выполнится ПОСЛЕ завершения основной функции,
		// даже если произошла паника.
		//
		// recover() перехватывает панику и возвращает значение, переданное в panic().
		// Если паники не было, recover() возвращает nil.
		defer func() {
			// Пытаемся перехватить панику
			if err := recover(); err != nil {
				requestID, _ := GetRequestIDFromContext(c)

				// ============================================
				// ПАНИКА ПЕРЕХВАЧЕНА - ОБРАБОТКА
				// ============================================

				// Получаем stack trace (информацию о том, где произошла паника)
				// debug.Stack() возвращает полный стек вызовов функций
				stack := debug.Stack()

				// Логируем панику с полными деталями
				// Это критическая ошибка, поэтому используем Error уровень
				event := logger.Error().
					Interface("panic", err).                 // Значение, переданное в panic()
					Str("stack", string(stack)).             // Полный stack trace
					Str("path", c.Request.URL.Path).         // Путь запроса
					Str("method", c.Request.Method).         // HTTP метод
					Str("client_ip", c.ClientIP()).          // IP адрес клиента
					Str("user_agent", c.Request.UserAgent()) // User-Agent браузера
				if requestID != "" {
					event.Str("request_id", requestID)
				}
				event.Msg("Panic recovered")

				// ============================================
				// ОТПРАВКА ОТВЕТА ПОЛЬЗОВАТЕЛЮ
				// ============================================
				// Преобразуем панику в AppError и отправляем пользователю
				// ВАЖНО: Не показываем детали паники пользователю (безопасность!)
				//
				// Создаём ошибку с общим сообщением
				// Детали паники остаются только в логах
				var panicErr error
				if errStr, ok := err.(string); ok {
					// Если panic() был вызван со строкой
					panicErr = fmt.Errorf("panic: %s", errStr)
				} else if errObj, ok := err.(error); ok {
					// Если panic() был вызван с error
					panicErr = errObj
				} else {
					// Если panic() был вызван с другим типом
					panicErr = fmt.Errorf("panic: %v", err)
				}

				// Используем нашу систему обработки ошибок
				// InternalError скроет детали от пользователя, но залогирует их
				appErr := errors.InternalError("An unexpected error occurred", panicErr)
				errors.HandleError(c, appErr)

				// ============================================
				// ПРЕРЫВАНИЕ ВЫПОЛНЕНИЯ
				// ============================================
				// c.Abort() уже вызван в HandleError,
				// но для ясности можно оставить комментарий
				// Дальнейшие middleware и handlers не будут выполнены
			}
		}()

		// ============================================
		// ПРОДОЛЖЕНИЕ ОБРАБОТКИ ЗАПРОСА
		// ============================================
		// Если паники не было, продолжаем выполнение цепочки middleware
		// Если была паника, выполнение прервётся в defer функции выше
		c.Next()
	}
}

// RecoveryMiddlewareWithStack создаёт Recovery middleware с опцией показа stack trace.
//
// ВАЖНО: Использовать только в режиме разработки (debug)!
// В продакшн никогда не показывайте stack trace пользователям!
//
// Параметры:
//   - showStack: показывать ли stack trace в ответе (true только для debug)
//
// Возвращает:
//   - gin.HandlerFunc: Recovery middleware
//
// Пример использования:
//
//	cfg := config.Get()
//	showStack := cfg.Server.Mode == "debug"
//	r.Use(middleware.RecoveryMiddlewareWithStack(showStack))
func RecoveryMiddlewareWithStack(showStack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := GetRequestIDFromContext(c)
				stack := debug.Stack()

				// Всегда логируем с полными деталями
				event := logger.Error().
					Interface("panic", err).
					Str("stack", string(stack)).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("client_ip", c.ClientIP())
				if requestID != "" {
					event.Str("request_id", requestID)
				}
				event.Msg("Panic recovered")

				// Формируем ответ
				var panicErr error
				if errStr, ok := err.(string); ok {
					panicErr = fmt.Errorf("panic: %s", errStr)
				} else if errObj, ok := err.(error); ok {
					panicErr = errObj
				} else {
					panicErr = fmt.Errorf("panic: %v", err)
				}

				// В режиме разработки можно показать больше информации
				if showStack {
					// В DEBUG режиме показываем stack trace в ответе
					// ВАЖНО: Только для разработки!
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "An unexpected error occurred",
						"success": false,
						"details": gin.H{
							"panic": fmt.Sprintf("%v", err),
							"stack": string(stack),
						},
					})
				} else {
					// В продакшн скрываем детали
					appErr := errors.InternalError("An unexpected error occurred", panicErr)
					errors.HandleError(c, appErr)
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}
