package middleware

import (
	"ctweb/internal/errors"
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware проверяет, является ли пользователь администратором.
//
// Что делает:
//   1. Получает пользователя из контекста (должен быть установлен в AuthMiddleware)
//   2. Проверяет, что пользователь принадлежит к группе администраторов (ID=1)
//   3. Проверяет, что группа администраторов активна
//   4. Если пользователь не администратор:
//      - HTML запросы → редирект на главную страницу или JSON ошибка 403
//      - API запросы → JSON ошибка 403 Forbidden
//
// Использование:
//   // В main.go
//   admin := r.Group("/admin")
//   admin.Use(middleware.AdminMiddleware())
//   admin.GET("/users", userController.ListUsers)
//
// Примечание:
//   Этот middleware должен использоваться ПОСЛЕ AuthMiddleware,
//   так как он полагается на наличие пользователя в контексте.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем пользователя из контекста (установлен в AuthMiddleware)
		user, exists := GetUserFromContext(c)
		if !exists {
			// Пользователь не найден в контексте
			// Это не должно произойти, если AuthMiddleware работает правильно
			logger.Error().
				Str("path", c.Request.URL.Path).
				Msg("User not found in context in AdminMiddleware")
			
			if isAPIRequest(c) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		// Проверяем, является ли пользователь администратором
		if !user.IsAdmin() {
			logger.Warn().
				Int("user_id", user.ID).
				Str("login", user.Login).
				Str("path", c.Request.URL.Path).
				Str("event", "admin_access_denied").
				Msg("Non-admin user attempted to access admin route")

			if isAPIRequest(c) {
				errors.HandleError(c, errors.ForbiddenError("Admin access required"))
			} else {
				// HTML запрос - редирект на главную страницу
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		// Проверяем, что группа администраторов активна
		// В PHP это проверяется в методе isAdmin()
		// Здесь проверяем дополнительно для безопасности
		if !isAdminGroupActive(user) {
			logger.Warn().
				Int("user_id", user.ID).
				Str("login", user.Login).
				Str("path", c.Request.URL.Path).
				Str("event", "admin_group_inactive").
				Msg("Admin group is inactive")

			if isAPIRequest(c) {
				errors.HandleError(c, errors.ForbiddenError("Admin group is inactive"))
			} else {
				c.Redirect(http.StatusFound, "/")
			}
			c.Abort()
			return
		}

		// Пользователь - администратор, продолжаем обработку
		c.Next()
	}
}

// isAdminGroupActive проверяет, активна ли группа администраторов.
//
// Параметры:
//   - user: пользователь для проверки
//
// Возвращает:
//   - bool: true если группа администраторов активна, false иначе
//
// Примечание:
//   Проверяет группу с ID=1 в базе данных.
//   Это дополнительная проверка безопасности.
func isAdminGroupActive(user *models.User) bool {
	// Проверяем, что пользователь принадлежит к группе администраторов
	if !user.HasGroup(1) {
		return false
	}

	// Проверяем активность группы администраторов в БД
	groupRepo := repositories.NewGroupRepository()
	group, err := groupRepo.FindByID(1)
	if err != nil {
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to check admin group status")
		return false
	}

	return group.IsActive()
}

// RequireAdmin проверяет права администратора в контроллере.
//
// Используется для проверки прав в контроллерах без использования middleware.
//
// Параметры:
//   - c: контекст Gin
//
// Возвращает:
//   - *models.User: данные пользователя-администратора
//   - bool: true если пользователь администратор, false иначе
//
// Использование:
//   admin, isAdmin := middleware.RequireAdmin(c)
//   if !isAdmin {
//       return // ошибка уже обработана
//   }
//   // Используем admin
func RequireAdmin(c *gin.Context) (*models.User, bool) {
	user, exists := GetUserFromContext(c)
	if !exists {
		if isAPIRequest(c) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	if !user.IsAdmin() {
		logger.Warn().
			Int("user_id", user.ID).
			Str("login", user.Login).
			Str("path", c.Request.URL.Path).
			Str("event", "admin_access_denied").
			Msg("Non-admin user attempted to access admin function")

		if isAPIRequest(c) {
			errors.HandleError(c, errors.ForbiddenError("Admin access required"))
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	// Проверяем активность группы администраторов
	if !isAdminGroupActive(user) {
		if isAPIRequest(c) {
			errors.HandleError(c, errors.ForbiddenError("Admin group is inactive"))
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	return user, true
}

// RequireGroup проверяет принадлежность пользователя к указанной группе.
//
// Параметры:
//   - c: контекст Gin
//   - groupID: ID группы для проверки
//
// Возвращает:
//   - *models.User: данные пользователя
//   - bool: true если пользователь принадлежит к группе, false иначе
//
// Использование:
//   user, hasGroup := middleware.RequireGroup(c, 2)
//   if !hasGroup {
//       return // ошибка уже обработана
//   }
func RequireGroup(c *gin.Context, groupID int) (*models.User, bool) {
	user, exists := GetUserFromContext(c)
	if !exists {
		if isAPIRequest(c) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	if !user.HasGroup(groupID) {
		logger.Warn().
			Int("user_id", user.ID).
			Str("login", user.Login).
			Int("required_group_id", groupID).
			Str("path", c.Request.URL.Path).
			Str("event", "group_access_denied").
			Msg("User does not belong to required group")

		if isAPIRequest(c) {
			errors.HandleError(c, errors.ForbiddenError("Group access required"))
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	return user, true
}

// RequireAnyGroup проверяет принадлежность пользователя хотя бы к одной из указанных групп.
//
// Параметры:
//   - c: контекст Gin
//   - groupIDs: список ID групп для проверки
//
// Возвращает:
//   - *models.User: данные пользователя
//   - bool: true если пользователь принадлежит хотя бы к одной группе, false иначе
//
// Использование:
//   user, hasAnyGroup := middleware.RequireAnyGroup(c, []int{1, 2})
//   if !hasAnyGroup {
//       return // ошибка уже обработана
//   }
func RequireAnyGroup(c *gin.Context, groupIDs []int) (*models.User, bool) {
	user, exists := GetUserFromContext(c)
	if !exists {
		if isAPIRequest(c) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	if !user.HasAnyGroup(groupIDs) {
		logger.Warn().
			Int("user_id", user.ID).
			Str("login", user.Login).
			Interface("required_group_ids", groupIDs).
			Str("path", c.Request.URL.Path).
			Str("event", "group_access_denied").
			Msg("User does not belong to any of the required groups")

		if isAPIRequest(c) {
			errors.HandleError(c, errors.ForbiddenError("Group access required"))
		} else {
			c.Redirect(http.StatusFound, "/")
		}
		c.Abort()
		return nil, false
	}

	return user, true
}

