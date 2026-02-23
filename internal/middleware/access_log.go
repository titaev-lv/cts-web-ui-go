package middleware

import (
	"ctweb/internal/config"
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
		meta := ExtractRequestMeta(c)
		userAgent := c.Request.UserAgent()

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		latencyMS := float64(latency.Microseconds()) / 1000.0
		cfg := config.Get()

		var userID any = nil
		if user, ok := GetUserFromContext(c); ok {
			userID = user.ID
		}

		requestID, _ := GetRequestIDFromContext(c)

		fields := []any{
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latencyMS,
			"ip", meta.RealIP,
			"user_agent", userAgent,
			"user_id", userID,
		}
		if cfg.Proxy.Enabled {
			fields = append(fields,
				"real_ip", meta.RealIP,
				"remote_addr", meta.RemoteAddr,
				"effective_scheme", meta.EffectiveScheme,
				"effective_host", meta.EffectiveHost,
				"trusted_proxy", meta.TrustedProxy,
			)
		}

		log.Info("HTTP access", fields...)
	}
}
