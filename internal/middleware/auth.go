package middleware

import (
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/session"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ContextKeyUser - ключ для хранения пользователя в контексте Gin
// Используется для доступа к данным пользователя в контроллерах
const ContextKeyUser = "user"

// AuthMiddleware проверяет аутентификацию пользователя через сессии и cookies.
//
// Что делает:
//  1. Проверяет, является ли маршрут публичным (не требует авторизации)
//  2. Пытается получить пользователя из сессии
//  3. Если сессии нет, пытается восстановить из cookies "Remember Me"
//  4. Если пользователь не авторизован:
//     - HTML запросы → редирект на /login
//     - API запросы → JSON ошибка 401
//  5. Сохраняет пользователя в контекст Gin для использования в контроллерах
//
// Публичные маршруты (не требуют авторизации):
//   - / - главная страница
//   - /login - страница входа
//   - /auth/login - обработка формы входа
//   - /assets/* - статические файлы
//
// Защищённые маршруты (требуют авторизации):
//   - Все остальные маршруты
//
// Использование в контроллерах:
//
//	user, exists := GetUserFromContext(c)
//	if !exists {
//	    // пользователь не найден (не должно произойти, т.к. middleware проверяет)
//	}
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ============================================
		// ПУБЛИЧНЫЕ МАРШРУТЫ (не требуют авторизации)
		// ============================================
		// ВСЕ страницы закрыты, кроме страницы входа и обработки входа
		public := map[string]bool{
			"/login":       true, // Страница входа
			"/auth/login":  true, // Обработка формы входа
			"/auth/logout": true, // Выход из системы (может быть вызван без авторизации)
			"/favicon.ico": true, // Favicon (браузеры запрашивают автоматически)
		}

		// Проверяем, является ли маршрут публичным
		if public[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Пропускаем статические файлы
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.Next()
			return
		}

		// ============================================
		// ЗАЩИЩЁННЫЕ МАРШРУТЫ (требуют авторизации)
		// ============================================
		authStart := time.Now()
		sm := session.GetSessionManager()

		// ШАГ 1: Пытаемся получить пользователя из сессии
		user, isAuth, err := sm.GetUser(c.Request)
		if err != nil {
			AddLatencyPart(c, "auth_middleware_ms", time.Since(authStart))
			logger.Error().
				Err(err).
				Str("path", c.Request.URL.Path).
				Msg("Failed to get user from session")

			// Ошибка при получении сессии - внутренняя ошибка сервера
			if isAPIRequest(c) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// ШАГ 2: Если сессии нет, пытаемся восстановить из cookies "Remember Me"
		if !isAuth {
			user, restored, err := sm.RestoreUserFromCookies(c.Request, c.Writer)
			if err != nil {
				AddLatencyPart(c, "auth_middleware_ms", time.Since(authStart))
				logger.Error().
					Err(err).
					Str("path", c.Request.URL.Path).
					Msg("Failed to restore user from cookies")

				// Ошибка при восстановлении - внутренняя ошибка сервера
				if isAPIRequest(c) {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Internal server error",
					})
				} else {
					c.Redirect(http.StatusFound, "/login")
				}
				c.Abort()
				return
			}

			if !restored {
				AddLatencyPart(c, "auth_middleware_ms", time.Since(authStart))
				// Пользователь не авторизован и не восстановлен из cookies
				// Редиректим на страницу входа или возвращаем JSON ошибку
				logger.Debug().
					Str("path", c.Request.URL.Path).
					Str("client_ip", c.ClientIP()).
					Msg("Unauthorized access attempt")

				if isAPIRequest(c) {
					c.JSON(http.StatusUnauthorized, gin.H{
						"error": "Unauthorized",
					})
				} else {
					// HTML запрос - редирект на страницу входа
					// Сохраняем URL для редиректа после входа (опционально)
					c.Redirect(http.StatusFound, "/login")
				}
				c.Abort()
				return
			}

			// Пользователь успешно восстановлен из cookies
			logger.Debug().
				Int("user_id", user.ID).
				Str("login", user.Login).
				Str("path", c.Request.URL.Path).
				Msg("User restored from Remember Me cookies")
		}

		// ШАГ 3: Проверяем, что пользователь активен
		if !user.IsActive() {
			AddLatencyPart(c, "auth_middleware_ms", time.Since(authStart))
			logger.Warn().
				Int("user_id", user.ID).
				Str("login", user.Login).
				Str("path", c.Request.URL.Path).
				Msg("Blocked user attempted to access protected route")

			// Очищаем сессию заблокированного пользователя
			sm.ClearUser(c.Request, c.Writer)

			if isAPIRequest(c) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "User is blocked",
				})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// ШАГ 4: Проверяем, что у пользователя есть активные группы
		if len(user.Groups) == 0 {
			AddLatencyPart(c, "auth_middleware_ms", time.Since(authStart))
			logger.Warn().
				Int("user_id", user.ID).
				Str("login", user.Login).
				Str("path", c.Request.URL.Path).
				Msg("User with no active groups attempted to access protected route")

			// Очищаем сессию пользователя без групп
			sm.ClearUser(c.Request, c.Writer)

			if isAPIRequest(c) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "User has no active groups",
				})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// ШАГ 5: Пользователь авторизован - сохраняем в контекст Gin
		// Это позволяет контроллерам получать данные пользователя через GetUserFromContext()
		c.Set(ContextKeyUser, user)
		AddLatencyPart(c, "auth_middleware_ms", time.Since(authStart))

		// Продолжаем обработку запроса
		c.Next()
	}
}

// isAPIRequest проверяет, является ли запрос API запросом.
//
// API запросы определяются по:
//   - Accept заголовок содержит "application/json"
//   - Или путь начинается с "/api/"
//
// Параметры:
//   - c: контекст Gin
//
// Возвращает:
//   - bool: true, если это API запрос
func isAPIRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		return true
	}

	return false
}

// GetUserFromContext получает пользователя из контекста Gin.
//
// Используется в контроллерах для получения данных текущего пользователя.
// Пользователь сохраняется в контекст в AuthMiddleware после успешной авторизации.
//
// Параметры:
//   - c: контекст Gin
//
// Возвращает:
//   - *models.User: данные пользователя (если авторизован)
//   - bool: true, если пользователь найден в контексте
//
// Использование:
//
//	user, exists := middleware.GetUserFromContext(c)
//	if !exists {
//	    // пользователь не найден (не должно произойти, т.к. middleware проверяет)
//	    c.JSON(500, gin.H{"error": "User not found in context"})
//	    return
//	}
//	// Используем данные пользователя
//	c.JSON(200, gin.H{"user_id": user.ID, "name": user.GetFullName()})
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	// Пытаемся получить пользователя из контекста
	value, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil, false
	}

	// Проверяем тип (должен быть *models.User)
	user, ok := value.(*models.User)
	if !ok {
		return nil, false
	}

	return user, true
}
