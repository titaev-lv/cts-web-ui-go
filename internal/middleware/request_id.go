package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader        = "X-Request-ID"
	ContextKeyRequestID    = "request_id"
	ContextKeyRequestStart = "request_start"
	ContextKeyLatencyParts = "latency_parts_ms"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(ContextKeyRequestID, requestID)
		c.Set(ContextKeyRequestStart, time.Now())
		c.Writer.Header().Set(RequestIDHeader, requestID)
		c.Next()
	}
}

func GetRequestIDFromContext(c *gin.Context) (string, bool) {
	if c == nil {
		return "", false
	}
	value, exists := c.Get(ContextKeyRequestID)
	if !exists {
		return "", false
	}
	requestID, ok := value.(string)
	if !ok || requestID == "" {
		return "", false
	}
	return requestID, true
}

func GetRequestStartFromContext(c *gin.Context) (time.Time, bool) {
	if c == nil {
		return time.Time{}, false
	}
	value, exists := c.Get(ContextKeyRequestStart)
	if !exists {
		return time.Time{}, false
	}
	startedAt, ok := value.(time.Time)
	if !ok || startedAt.IsZero() {
		return time.Time{}, false
	}
	return startedAt, true
}

func AddLatencyPart(c *gin.Context, key string, d time.Duration) {
	AddLatencyPartMS(c, key, float64(d.Microseconds())/1000.0)
}

func AddLatencyPartMS(c *gin.Context, key string, value float64) {
	if c == nil || key == "" {
		return
	}
	if value < 0 {
		value = 0
	}

	parts, _ := GetLatencyParts(c)
	if parts == nil {
		parts = map[string]float64{}
	}
	parts[key] = value
	c.Set(ContextKeyLatencyParts, parts)
}

func GetLatencyParts(c *gin.Context) (map[string]float64, bool) {
	if c == nil {
		return nil, false
	}
	value, exists := c.Get(ContextKeyLatencyParts)
	if !exists {
		return nil, false
	}
	parts, ok := value.(map[string]float64)
	if !ok {
		return nil, false
	}
	return parts, true
}

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
}
