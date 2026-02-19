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
	RequestIDHeader     = "X-Request-ID"
	ContextKeyRequestID = "request_id"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(ContextKeyRequestID, requestID)
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

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
}
