package middleware

import (
	"strings"
	"time"

	"ctweb/internal/logger"

	"github.com/gin-gonic/gin"
)

// AuditLogMiddleware writes security/administrative audit events.
func AuditLogMiddleware() gin.HandlerFunc {
	log := logger.Audit()
	return func(c *gin.Context) {
		if !shouldAudit(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

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
		requestID, _ := GetRequestIDFromContext(c)

		action, resourceType := inferAuditAction(path, method)

		var userID any = nil
		var userLogin any = nil
		if user, ok := GetUserFromContext(c); ok {
			userID = user.ID
			userLogin = user.Login
		}

		result := "success"
		if status >= 400 {
			result = "failure"
		}

		event := log.Info
		if status >= 400 {
			event = log.Warn
		}

		event("audit",
			"module", "audit",
			"event_type", "audit",
			"action", action,
			"resource_type", resourceType,
			"method", method,
			"path", path,
			"status", status,
			"result", result,
			"latency_ms", latency.Milliseconds(),
			"ip", clientIP,
			"user_agent", userAgent,
			"request_id", requestID,
			"user_id", userID,
			"user_login", userLogin,
		)
	}
}

func shouldAudit(method, path string) bool {
	m := strings.ToUpper(method)
	p := strings.ToLower(path)

	if strings.HasPrefix(p, "/assets/") || p == "/favicon.ico" {
		return false
	}

	if p == "/auth/login" || p == "/auth/logout" {
		return true
	}

	if strings.Contains(p, "/ajax_get") {
		return false
	}

	switch m {
	case "PUT", "PATCH", "DELETE":
		return true
	case "POST":
		return true
	default:
		return false
	}
}

func inferAuditAction(path string, method string) (string, string) {
	p := strings.ToLower(path)
	m := strings.ToUpper(method)

	resourceType := "web_ui"
	if strings.HasPrefix(p, "/users") {
		resourceType = "user"
	} else if strings.HasPrefix(p, "/groups") {
		resourceType = "group"
	} else if strings.HasPrefix(p, "/exchange_manage") {
		resourceType = "exchange"
	} else if strings.HasPrefix(p, "/exchange_accounts") {
		resourceType = "exchange_account"
	} else if strings.HasPrefix(p, "/auth") {
		resourceType = "auth"
	}

	action := m + "_" + resourceType
	if strings.Contains(p, "ajax_create") {
		action = "CREATE_" + resourceType
	} else if strings.Contains(p, "ajax_edit") {
		action = "UPDATE_" + resourceType
	} else if strings.Contains(p, "ajax_delete") {
		action = "DELETE_" + resourceType
	} else if p == "/auth/login" {
		action = "LOGIN"
	} else if p == "/auth/logout" {
		action = "LOGOUT"
	}

	return action, resourceType
}
