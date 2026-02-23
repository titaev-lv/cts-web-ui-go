package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
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

type contextRequestIDKey struct{}

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

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, contextRequestIDKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	requestID, ok := ctx.Value(contextRequestIDKey{}).(string)
	if !ok {
		return "", false
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return "", false
	}
	return requestID, true
}

func SetRequestIDHeaderFromContext(req *http.Request) {
	if req == nil {
		return
	}
	if requestID, ok := RequestIDFromContext(req.Context()); ok {
		req.Header.Set(RequestIDHeader, requestID)
	}
}

func NewRequestWithRequestID(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	SetRequestIDHeaderFromContext(req)
	return req, nil
}

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
}
