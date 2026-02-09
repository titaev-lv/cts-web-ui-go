// Package controllers содержит HTTP-контроллеры для работы с биржами.
package controllers

import (
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"ctweb/internal/services"
	"ctweb/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExchangeController обрабатывает запросы, связанные с биржами.
type ExchangeController struct {
	service *services.ExchangeService
}

// NewExchangeController создаёт новый экземпляр ExchangeController.
func NewExchangeController() *ExchangeController {
	return &ExchangeController{
		service: services.NewExchangeService(),
	}
}

// List отображает страницу управления биржами.
func (ec *ExchangeController) List(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "exchange/index.html", gin.H{
		"Title": "Exchange Manage",
		"User":  user.(*models.User),
	})
}

// AjaxGetExchanges отдаёт данные для DataTables (список бирж).
func (ec *ExchangeController) AjaxGetExchanges(c *gin.Context) {
	req := utils.ParseDataTablesRequest(c)
	repoReq := utils.ConvertToExchangeRepositoryRequest(req)

	repoResp, err := ec.service.ExchangeRepo().FindAllWithPagination(repoReq)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get exchanges")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load exchanges"})
		return
	}

	resp := utils.ConvertExchangeResponseToDataTablesFormat(req.Draw, repoResp)
	c.JSON(http.StatusOK, resp)
}

// AjaxGetExchangeByID возвращает данные биржи для модальной формы редактирования.
func (ec *ExchangeController) AjaxGetExchangeByID(c *gin.Context) {
	idStr := c.PostForm("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ex, err := ec.service.ExchangeRepo().FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "exchange not found"})
		return
	}

	status := "Blocked"
	if ex.Active {
		status = "Active"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            ex.ID,
		"name":          ex.Name,
		"status":        status,
		"url":           ex.URL,
		"base_url":      ex.BaseURL,
		"websocket_url": ex.WebsocketURL,
		"class":         ex.ClassToFactory,
		"description":   ex.Description,
	})
}

// AjaxCreateExchange создаёт новую биржу.
func (ec *ExchangeController) AjaxCreateExchange(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)
	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	name := c.PostForm("create_exchange_name")
	url := c.PostForm("create_exchange_url")
	baseURL := c.PostForm("create_exchange_base_url")
	websocketURL := c.PostForm("create_exchange_websocket_url")
	classToFactory := c.PostForm("create_exchange_class")
	status := c.PostForm("create_exchange_status")
	description := c.PostForm("create_exchange_description")
	if description == "" {
		description = c.PostForm("create_exchange_desc")
	}

	var descPtr *string
	if description != "" {
		trimmed := strings.TrimSpace(description)
		descPtr = &trimmed
	}

	var websocketURLPtr *string
	if websocketURL != "" {
		trimmed := strings.TrimSpace(websocketURL)
		websocketURLPtr = &trimmed
	}

	id, err := ec.service.CreateExchange(name, url, baseURL, classToFactory, status, descPtr, websocketURLPtr, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "id": id})
}

// AjaxEditExchange обновляет биржу.
func (ec *ExchangeController) AjaxEditExchange(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)
	if !user.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	idStr := c.PostForm("edit_exchange_id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	name := c.PostForm("edit_exchange_name")
	url := c.PostForm("edit_exchange_url")
	baseURL := c.PostForm("edit_exchange_base_url")
	websocketURL := c.PostForm("edit_exchange_websocket_url")
	classToFactory := c.PostForm("edit_exchange_class")
	status := c.PostForm("edit_exchange_status")
	description := c.PostForm("edit_exchange_description")
	if description == "" {
		description = c.PostForm("edit_exchange_desc")
	}

	var descPtr *string
	if description != "" {
		trimmed := strings.TrimSpace(description)
		descPtr = &trimmed
	}

	var websocketURLPtr *string
	if websocketURL != "" {
		trimmed := strings.TrimSpace(websocketURL)
		websocketURLPtr = &trimmed
	}

	if err := ec.service.UpdateExchange(id, name, url, baseURL, classToFactory, status, descPtr, websocketURLPtr, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ExchangeRepo позволяет переиспользовать репозиторий (для ajax_get_exchanges).
func (ec *ExchangeController) ExchangeRepo() *repositories.ExchangeRepository {
	return ec.service.ExchangeRepo()
}
