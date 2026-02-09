// Package middleware - Rate Limiting для защиты от брутфорса и DDoS атак.
package middleware

import (
	"ctweb/internal/config"
	"ctweb/internal/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"                      // Rate limiting библиотека
	"github.com/ulule/limiter/v3/drivers/store/memory" // Хранение лимитов в памяти
)

// RateLimitMiddleware создаёт middleware для ограничения количества запросов.
//
// Что такое Rate Limiting:
//
//	Rate Limiting ограничивает количество запросов от одного IP адреса
//	за определённый период времени. Это защищает от:
//	- Брутфорс атак (множественные попытки входа)
//	- DDoS атак (перегрузка сервера запросами)
//	- Злоупотребления API
//
// Как работает:
//  1. Считает количество запросов от каждого IP адреса
//  2. Если лимит превышен, возвращает 429 Too Many Requests
//  3. Счётчик сбрасывается через определённое время
//
// Параметры:
//   - limit: количество разрешённых запросов (например, "5" для 5 запросов)
//   - period: период времени (например, "1-M" для 1 минуты)
//   - message: сообщение об ошибке при превышении лимита
//
// Формат period:
//   - "1-S" = 1 секунда
//   - "1-M" = 1 минута
//   - "1-H" = 1 час
//   - "1-D" = 1 день
//
// Возвращает:
//   - gin.HandlerFunc: rate limit middleware
//
// Пример использования:
//
//	// Ограничение: 5 запросов в минуту
//	loginRateLimit := middleware.RateLimitMiddleware("5", "1-M", "Too many login attempts")
//	auth.POST("/login", loginRateLimit, loginHandler)
//
//	// Ограничение: 100 запросов в секунду
//	apiRateLimit := middleware.RateLimitMiddleware("100", "1-S", "API rate limit exceeded")
//	api := r.Group("/api", apiRateLimit)
func RateLimitMiddleware(limit, period, message string) gin.HandlerFunc {
	// Создаём rate limiter с указанным лимитом и периодом
	// Формат limit: "5-M" означает 5 запросов в минуту
	// Если period пустой, то limit уже содержит полный формат (например, "5-M")
	// Если period указан, то формируем формат из limit и period
	limitStr := limit
	if period != "" {
		limitStr = limit + "-" + period
	}
	rate, err := limiter.NewRateFromFormatted(limitStr)
	if err != nil {
		// Если не удалось создать rate limiter, логируем ошибку
		// и возвращаем middleware, который ничего не делает
		logger.Error().
			Err(err).
			Str("limit", limit).
			Str("period", period).
			Msg("Failed to create rate limiter, rate limiting disabled")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Создаём хранилище для счётчиков в памяти
	// В продакшн лучше использовать Redis для распределённых систем
	// memory.NewStore() возвращает только store, без ошибки
	store := memory.NewStore()

	// Создаём экземпляр limiter
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		// Получаем IP адрес клиента
		// X-Real-IP и X-Forwarded-For используются, если приложение за прокси (nginx)
		clientIP := c.ClientIP()
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		} else if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
			clientIP = realIP
		}

		// Получаем контекст для этого IP адреса
		context, err := instance.Get(c.Request.Context(), clientIP)
		if err != nil {
			// Если ошибка при получении контекста, пропускаем запрос
			// (лучше пропустить, чем заблокировать всех)
			logger.Warn().
				Err(err).
				Str("client_ip", clientIP).
				Msg("Rate limiter error, allowing request")
			c.Next()
			return
		}

		// Устанавливаем заголовки с информацией о лимите
		// Это полезно для клиентов, чтобы они знали свои лимиты
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		// Проверяем, не превышен ли лимит
		if context.Reached {
			// Лимит превышен, возвращаем 429 Too Many Requests
			logger.Warn().
				Str("client_ip", clientIP).
				Str("path", c.Request.URL.Path).
				Int64("limit", context.Limit).
				Msg("Rate limit exceeded")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   message,
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Лимит не превышен, продолжаем обработку
		c.Next()
	}
}

// LoginRateLimitMiddleware создаёт rate limiter для страницы входа.
//
// Использует настройки из конфигурации (rate_limit_login).
//
// Возвращает:
//   - gin.HandlerFunc: rate limit middleware для логина
//
// Пример использования:
//
//	auth.POST("/login", middleware.LoginRateLimitMiddleware(), loginHandler)
func LoginRateLimitMiddleware() gin.HandlerFunc {
	cfg := config.Get()
	// Используем настройку из конфигурации: X запросов в минуту
	// Формат: "5-M" означает 5 запросов в минуту
	limit := strconv.Itoa(cfg.Security.RateLimitLogin) + "-M"
	return RateLimitMiddleware(limit, "", "Too many login attempts. Please try again later.")
}

// APIRateLimitMiddleware создаёт rate limiter для API endpoints.
//
// Использует настройки из конфигурации (rate_limit_api).
//
// Возвращает:
//   - gin.HandlerFunc: rate limit middleware для API
//
// Пример использования:
//
//	api := r.Group("/api", middleware.APIRateLimitMiddleware())
func APIRateLimitMiddleware() gin.HandlerFunc {
	cfg := config.Get()
	// Используем настройку из конфигурации: X запросов в секунду
	// Формат: "100-S" означает 100 запросов в секунду
	limit := strconv.Itoa(cfg.Security.RateLimitAPI) + "-S"
	return RateLimitMiddleware(limit, "", "API rate limit exceeded. Please slow down.")
}
