package middleware

import (
	"time"

	"ctweb/internal/logger"

	"github.com/gin-gonic/gin"
)

// AccessLogMiddleware writes access logs for HTTP requests.
func AccessLogMiddleware() gin.HandlerFunc {
	log := logger.Access()
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if rawQuery := c.Request.URL.RawQuery; rawQuery != "" {
			path = path + "?" + rawQuery
		}

		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		var userID any = nil
		if user, ok := GetUserFromContext(c); ok {
			userID = user.ID
		}

		requestID, _ := GetRequestIDFromContext(c)

		log.Info("HTTP access",
			"module", "access",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", clientIP,
			"user_agent", userAgent,
			"user_id", userID,
		)
	}
}
