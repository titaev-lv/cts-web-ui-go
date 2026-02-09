// Package controllers содержит HTTP handlers (контроллеры) для обработки запросов.
// Контроллеры получают запросы, вызывают сервисы и возвращают ответы.
package controllers

import (
	"ctweb/internal/logger"      // Система логирования
	"ctweb/internal/middleware"   // Middleware для получения пользователя
	"ctweb/internal/models"       // Модели данных
	"ctweb/internal/repositories" // Репозитории для работы с БД
	"ctweb/internal/utils"        // Утилиты (DataTables парсинг, валидация)
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GroupController - контроллер для работы с группами пользователей.
// Содержит методы для обработки HTTP запросов, связанных с группами.
type GroupController struct{}

// NewGroupController создаёт новый экземпляр GroupController.
//
// Возвращает:
//   - *GroupController: новый контроллер
func NewGroupController() *GroupController {
	return &GroupController{}
}

// AjaxGetGroups обрабатывает AJAX запрос от DataTables для получения списка групп.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//   1. Проверяет, что пользователь - администратор
//   2. Парсит параметры DataTables запроса
//   3. Вызывает репозиторий для получения данных
//   4. Форматирует ответ в формате DataTables (aaData)
//
// Формат запроса:
//   POST /groups/ajax_get_groups
//   Параметры DataTables (draw, start, length, search[value], order[0][column], columns[i][data] и т.д.)
//
// Формат ответа:
//   {
//     "draw": 1,
//     "recordsTotal": 10,
//     "recordsFiltered": 5,
//     "aaData": [
//       {
//         "chbx": "",
//         "DT_RowId": "row_1",
//         "id": 1,
//         "name": "Administrators",
//         "status": "Active",
//         "description": "System administrators"
//       }
//     ]
//   }
func (g *GroupController) AjaxGetGroups(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров DataTables запроса
	// ============================================
	req := utils.ParseDataTablesRequest(c)
	
	// Логирование для отладки (только в debug режиме)
	if len(req.Order) > 0 {
		logger.Debug().
			Int("draw", req.Draw).
			Int("start", req.Start).
			Int("length", req.Length).
			Str("search", req.Search).
			Int("order_count", len(req.Order)).
			Int("order_column", req.Order[0].Column).
			Str("order_dir", req.Order[0].Dir).
			Int("columns_count", len(req.Columns)).
			Msg("DataTables request parsed for groups")
	} else {
		logger.Debug().
			Int("draw", req.Draw).
			Int("start", req.Start).
			Int("length", req.Length).
			Str("search", req.Search).
			Int("order_count", 0).
			Int("columns_count", len(req.Columns)).
			Msg("DataTables request parsed for groups (no order)")
	}
	
	// Логируем значения поиска по колонкам
	for i, col := range req.Columns {
		if col.Search.Value != "" {
			logger.Debug().
				Int("column_index", i).
				Str("column_data", col.Data).
				Str("search_value", col.Search.Value).
				Bool("searchable", col.Searchable).
				Msg("Column search value for groups")
		}
	}

	// ============================================
	// ШАГ 3: Получение данных из репозитория
	// ============================================
	groupRepo := repositories.NewGroupRepository()
	
	// Конвертируем запрос в формат репозитория
	groupReq := utils.ConvertToGroupRepositoryRequest(req)
	
	// Получаем данные
	repoResponse, err := groupRepo.FindAllWithPagination(groupReq)
	if err != nil {
		logger.Error().
			Err(err).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to get groups list for DataTables")
		
		// Возвращаем пустой ответ в случае ошибки (как в PHP)
		c.JSON(http.StatusOK, gin.H{
			"recordsTotal":    0,
			"recordsFiltered": 0,
			"aaData":          []interface{}{},
		})
		return
	}

	// ============================================
	// ШАГ 4: Форматирование ответа в формат DataTables
	// ============================================
	response := utils.ConvertGroupResponseToDataTablesFormat(req.Draw, repoResponse)

	// ============================================
	// ШАГ 5: Возврат ответа
	// ============================================
	c.JSON(http.StatusOK, response)
}

// List отображает страницу управления группами.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//   1. Проверяет, что пользователь - администратор
//   2. Рендерит HTML шаблон страницы групп
//
// Формат запроса:
//   GET /groups/
//
// Шаблон:
//   - groups/index.html (если будет создан)
func (g *GroupController) List(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ============================================
	// ШАГ 2: Рендеринг страницы
	// ============================================
	c.HTML(http.StatusOK, "groups/index.html", gin.H{
		"Title": "Groups Management",
		"User":  user,
	})
}

// AjaxGetGroupById обрабатывает AJAX запрос для получения данных группы по ID.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//   1. Проверяет, что пользователь - администратор
//   2. Получает ID группы из параметров запроса
//   3. Загружает группу из БД
//   4. Форматирует ответ в формате PHP (lowercase поля)
//
// Формат запроса:
//   GET /groups/ajax_getid_group?id=1
//   или
//   POST /groups/ajax_getid_group с параметром id
//
// Формат ответа:
//   {
//     "error": false,
//     "success": true,
//     "data": {
//       "id": "1",
//       "name": "Administrators",
//       "description": "System administrators",
//       "status": "enable"
//     }
//   }
func (g *GroupController) AjaxGetGroupById(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Получение ID группы из параметров
	// ============================================
	// Поддерживаем оба формата: GET (query) и POST (form)
	var groupIDStr string
	if c.Request.Method == "GET" {
		groupIDStr = c.Query("id")
	} else {
		groupIDStr = c.PostForm("id")
	}

	if groupIDStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Group ID is required",
			"success": false,
		})
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Invalid group ID",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 3: Получение группы из БД
	// ============================================
	groupRepo := repositories.NewGroupRepository()
	targetGroup, err := groupRepo.FindByID(groupID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("group_id", groupID).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to get group by ID")

		c.JSON(http.StatusOK, gin.H{
			"error":   "DB ERROR: Error select group",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Форматирование ответа (как в PHP - lowercase поля)
	// ============================================
	status := utils.BoolToStatus(targetGroup.Active)

	responseData := gin.H{
		"id":          strconv.Itoa(targetGroup.ID),
		"name":        targetGroup.Name,
		"description": targetGroup.Description,
		"status":      status,
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
		"data":    responseData,
	})
}

// AjaxCreateGroup обрабатывает AJAX запрос для создания новой группы.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//   1. Проверяет, что пользователь - администратор
//   2. Парсит и валидирует все поля формы
//   3. Проверяет уникальность имени группы
//   4. Создаёт группу в БД
//
// Формат запроса:
//   POST /groups/ajax_create_group
//   Параметры:
//     - create_group_name: название группы (обязательно)
//     - create_group_description: описание группы (опционально)
//     - create_group_status: статус ("enable" или "disable") (обязательно)
//
// Формат ответа:
//   {
//     "error": false,
//     "success": true
//   }
//   или
//   {
//     "error": "Error message",
//     "success": false
//   }
func (g *GroupController) AjaxCreateGroup(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров формы
	// ============================================
	groupName := strings.TrimSpace(c.PostForm("create_group_name"))
	groupDescription := c.PostForm("create_group_description")
	status := c.PostForm("create_group_status")

	// ============================================
	// ШАГ 3: Валидация обязательных полей
	// ============================================
	var errorMsg string

	if groupName == "" {
		errorMsg = "Field \"Group Name\" is empty"
	} else if status == "" {
		errorMsg = "Field \"Status\" is empty"
	}

	if errorMsg != "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Валидация статуса
	// ============================================
	if err := utils.ValidateStatus(status); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 5: Проверка уникальности имени группы
	// ============================================
	groupRepo := repositories.NewGroupRepository()
	existsName, err := groupRepo.ExistsByName(groupName)
	if err != nil {
		logger.Error().
			Err(err).
			Str("group_name", groupName).
			Msg("Failed to check group name uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsName {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Group name already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6: Создание группы
	// ============================================
	newGroup := &models.Group{
		Name:        groupName,
		Description: groupDescription,
		Active:      utils.StatusToBool(status),
	}

	createdGroupID, err := groupRepo.Create(newGroup, user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("group_name", groupName).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to create group")

		c.JSON(http.StatusOK, gin.H{
			"error":   "Error create group",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7: Логирование успешного создания
	// ============================================
	logger.Info().
		Int("created_group_id", createdGroupID).
		Str("group_name", groupName).
		Int("creator_user_id", user.ID).
		Str("client_ip", c.ClientIP()).
		Msg("Group created successfully")

	// ============================================
	// ШАГ 8: Возврат успешного ответа
	// ============================================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}

// AjaxEditGroup обрабатывает AJAX запрос для редактирования группы.
//
// Параметры:
//   - c: контекст Gin
//
// Что делает:
//   1. Проверяет, что пользователь - администратор
//   2. Парсит и валидирует все поля формы
//   3. Проверяет уникальность имени группы (исключая текущую группу)
//   4. Обновляет группу в БД
//
// Формат запроса:
//   POST /groups/ajax_edit_group
//   Параметры:
//     - edit_group_id: ID группы (обязательно)
//     - edit_group_name: название группы (обязательно)
//     - edit_group_description: описание группы (опционально)
//     - edit_group_status: статус ("enable" или "disable") (обязательно)
//
// Формат ответа:
//   {
//     "error": false,
//     "success": true
//   }
//   или
//   {
//     "error": "Error message",
//     "success": false
//   }
func (g *GroupController) AjaxEditGroup(c *gin.Context) {
	// ============================================
	// ШАГ 1: Проверка прав доступа (только администратор)
	// ============================================
	user, exists := middleware.GetUserFromContext(c)
	if !exists || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Unauthorized",
			"success": false,
		})
		return
	}

	if !user.IsAdmin() {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Access deny",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 2: Парсинг параметров формы
	// ============================================
	groupIDStr := c.PostForm("edit_group_id")
	groupName := strings.TrimSpace(c.PostForm("edit_group_name"))
	groupDescription := c.PostForm("edit_group_description")
	status := c.PostForm("edit_group_status")

	// ============================================
	// ШАГ 3: Валидация обязательных полей
	// ============================================
	var errorMsg string

	if groupIDStr == "" {
		errorMsg = "Field \"Group ID\" is empty"
	} else if groupName == "" {
		errorMsg = "Field \"Group Name\" is empty"
	} else if status == "" {
		errorMsg = "Field \"Status\" is empty"
	}

	if errorMsg != "" {
		c.JSON(http.StatusOK, gin.H{
			"error":   errorMsg,
			"success": false,
		})
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Invalid group ID",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 4: Проверка уникальности имени группы (исключая текущую группу)
	// ============================================
	groupRepo := repositories.NewGroupRepository()
	existsName, err := groupRepo.ExistsByNameExcludingID(groupName, groupID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("group_name", groupName).
			Int("group_id", groupID).
			Msg("Failed to check group name uniqueness")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Database error",
			"success": false,
		})
		return
	}
	if existsName {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Group name already exist",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 5: Валидация статуса
	// ============================================
	if err := utils.ValidateStatus(status); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 6: Получение существующей группы
	// ============================================
	existingGroup, err := groupRepo.FindByID(groupID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("group_id", groupID).
			Msg("Failed to find group for update")
		c.JSON(http.StatusOK, gin.H{
			"error":   "Group not found",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 7: Обновление данных группы
	// ============================================
	existingGroup.Name = groupName
	existingGroup.Description = groupDescription
	existingGroup.Active = utils.StatusToBool(status)

	err = groupRepo.Update(existingGroup, user.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Int("group_id", groupID).
			Str("client_ip", c.ClientIP()).
			Msg("Failed to update group")

		c.JSON(http.StatusOK, gin.H{
			"error":   "Error edit group",
			"success": false,
		})
		return
	}

	// ============================================
	// ШАГ 8: Логирование успешного обновления
	// ============================================
	logger.Info().
		Int("updated_group_id", groupID).
		Str("group_name", groupName).
		Int("editor_user_id", user.ID).
		Str("client_ip", c.ClientIP()).
		Msg("Group updated successfully")

	// ============================================
	// ШАГ 9: Возврат успешного ответа
	// ============================================
	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"success": true,
	})
}

