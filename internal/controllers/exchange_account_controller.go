// Package controllers содержит HTTP-контроллеры для работы с аккаунтами бирж.
package controllers

import (
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/services"
	"ctweb/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExchangeAccountController обрабатывает запросы, связанные с аккаунтами бирж.
type ExchangeAccountController struct {
	service *services.ExchangeService
}

// NewExchangeAccountController создаёт новый экземпляр ExchangeAccountController.
func NewExchangeAccountController() *ExchangeAccountController {
	return &ExchangeAccountController{
		service: services.NewExchangeService(),
	}
}

// List отображает страницу аккаунтов бирж (шаблон добавим в 4.6).
func (eac *ExchangeAccountController) List(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "exchange_accounts/index.html", gin.H{
		"Title": "Exchange Accounts",
		"User":  user.(*models.User),
	})
}

// AjaxGetAccounts отдаёт данные аккаунтов пользователя для DataTables.
// Пока без серверной фильтрации, но с пагинацией (start/length) и draw.
func (eac *ExchangeAccountController) AjaxGetAccounts(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	req := utils.ParseDataTablesRequest(c)

	// Загружаем все аккаунты пользователя и мапу бирж (ID -> Name).
	accounts, err := eac.service.AccountRepo().FindAllByUser(user.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", user.ID).Msg("failed to load exchange accounts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load exchange accounts"})
		return
	}

	exchanges, err := eac.service.ExchangeRepo().FindAllActive()
	if err != nil {
		logger.Error().Err(err).Msg("failed to load exchanges for accounts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load exchanges"})
		return
	}
	exName := make(map[int]string)
	for _, ex := range exchanges {
		exName[ex.ID] = ex.Name
	}

	recordsTotal := len(accounts)
	recordsFiltered := recordsTotal // пока без серверной фильтрации

	// Пагинация вручную
	start := req.Start
	if start < 0 {
		start = 0
	}
	end := start + req.Length
	if end > len(accounts) {
		end = len(accounts)
	}
	if start > end {
		start = end
	}
	page := accounts[start:end]

	aaData := make([]map[string]interface{}, len(page))
	for i, acc := range page {
		status := "Blocked"
		if acc.Active {
			status = "Active"
		}
		aaData[i] = map[string]interface{}{
			"chbx":          "",
			"DT_RowId":      "row_" + strconv.Itoa(acc.ID),
			"id":            acc.ID,
			"exchange_name": exName[acc.ExID],
			"exchange_id":   acc.ExID,
			"account_name":  acc.AccountName,
			"priority":      acc.Priority,
			"status":        status,
			"api_key":       acc.ApiKey,
			"note":          acc.Note,
		}
	}

	c.JSON(http.StatusOK, utils.DataTablesResponse{
		Draw:            req.Draw,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		AAData:          aaData,
	})
}

// AjaxGetAccountByID возвращает данные аккаунта по ID.
func (eac *ExchangeAccountController) AjaxGetAccountByID(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	idStr := c.PostForm("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	acc, err := eac.service.AccountRepo().FindByID(id, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	status := "Blocked"
	if acc.Active {
		status = "Active"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           acc.ID,
		"exchange_id":  acc.ExID,
		"account_name": acc.AccountName,
		"priority":     acc.Priority,
		"status":       status,
		"api_key":      acc.ApiKey,
		"secret_key":   acc.SecretKey,
		"add_key":      acc.AddKey,
		"note":         acc.Note,
	})
}

// AjaxCreateAccount создаёт новый аккаунт биржи.
func (eac *ExchangeAccountController) AjaxCreateAccount(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	exidStr := c.PostForm("create_exchange_account_exid")
	accountName := c.PostForm("create_exchange_account_account_name")
	status := c.PostForm("create_exchange_account_status")
	priorityStr := c.PostForm("create_exchange_account_priority")
	apiKey := c.PostForm("create_exchange_account_api_key")
	secretKey := c.PostForm("create_exchange_account_secret_key")
	addKey := c.PostForm("create_exchange_account_add_key")
	note := c.PostForm("create_exchange_account_note")

	exid, err := strconv.Atoi(exidStr)
	if err != nil || exid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange id"})
		return
	}
	if err := eac.service.ValidateExchangeExists(exid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := 0
	if strings.TrimSpace(priorityStr) != "" {
		if p, err := strconv.Atoi(priorityStr); err == nil {
			priority = p
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority"})
			return
		}
	}

	id, err := eac.service.CreateExchangeAccount(user.ID, exid, accountName, status, priority, apiKey, secretKey, addKey, note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "id": id})
}

// AjaxEditAccount обновляет аккаунт биржи.
func (eac *ExchangeAccountController) AjaxEditAccount(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	idStr := c.PostForm("edit_exchange_account_id")
	exidStr := c.PostForm("edit_exchange_account_exid")
	accountName := c.PostForm("edit_exchange_account_account_name")
	status := c.PostForm("edit_exchange_account_status")
	priorityStr := c.PostForm("edit_exchange_account_priority")
	apiKey := c.PostForm("edit_exchange_account_api_key")
	secretKey := c.PostForm("edit_exchange_account_secret_key")
	addKey := c.PostForm("edit_exchange_account_add_key")
	note := c.PostForm("edit_exchange_account_note")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	exid, err := strconv.Atoi(exidStr)
	if err != nil || exid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange id"})
		return
	}
	if err := eac.service.ValidateExchangeExists(exid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priority := 0
	if strings.TrimSpace(priorityStr) != "" {
		if p, err := strconv.Atoi(priorityStr); err == nil {
			priority = p
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority"})
			return
		}
	}

	if err := eac.service.UpdateExchangeAccount(id, user.ID, exid, accountName, status, priority, apiKey, secretKey, addKey, note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
