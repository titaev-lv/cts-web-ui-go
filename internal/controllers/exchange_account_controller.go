// Package controllers содержит HTTP-контроллеры для работы с аккаунтами бирж.
package controllers

import (
	"ctweb/internal/logger"
	"ctweb/internal/models"
	"ctweb/internal/services"
	"ctweb/internal/utils"
	"net/http"
	"sort"
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

	exchanges, err := eac.service.ExchangeRepo().FindAllNamesWithStatus()
	if err != nil {
		logger.Error().Err(err).Msg("failed to load exchanges for account forms")
		exchanges = []*models.Exchange{}
	}

	c.HTML(http.StatusOK, "exchange_accounts/index.html", gin.H{
		"Title":     "Exchange Accounts",
		"User":      user.(*models.User),
		"Exchanges": exchanges,
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

	// Загружаем все аккаунты пользователя (с exchange name через LEFT JOIN).
	accounts, err := eac.service.AccountRepo().FindAllByUser(user.ID)
	if err != nil {
		logger.Error().Err(err).Int("user_id", user.ID).Msg("failed to load exchange accounts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load exchange accounts"})
		return
	}

	recordsTotal := len(accounts)

	getStatus := func(acc *models.ExchangeAccount) string {
		if acc.Active && acc.ExchangeActive {
			return "Active"
		}
		return "Blocked"
	}

	getExchangeName := func(acc *models.ExchangeAccount) string {
		exchangeName := strings.TrimSpace(acc.ExchangeName)
		if exchangeName == "" {
			return strconv.Itoa(acc.ExID)
		}
		return exchangeName
	}

	getNote := func(acc *models.ExchangeAccount) string {
		if acc.Note == nil {
			return ""
		}
		return strings.TrimSpace(*acc.Note)
	}

	matchesColumn := func(acc *models.ExchangeAccount, colData, searchValue string) bool {
		needle := strings.ToLower(strings.TrimSpace(searchValue))
		if needle == "" {
			return true
		}

		switch colData {
		case "id":
			return strings.Contains(strings.ToLower(strconv.Itoa(acc.ID)), needle)
		case "exchange_name":
			return strings.Contains(strings.ToLower(getExchangeName(acc)), needle)
		case "account_name":
			return strings.Contains(strings.ToLower(acc.AccountName), needle)
		case "note":
			return strings.Contains(strings.ToLower(getNote(acc)), needle)
		case "priority":
			return strings.Contains(strings.ToLower(strconv.Itoa(acc.Priority)), needle)
		case "status":
			return strings.Contains(strings.ToLower(getStatus(acc)), needle)
		default:
			return true
		}
	}

	filtered := make([]*models.ExchangeAccount, 0, len(accounts))
	globalNeedle := strings.ToLower(strings.TrimSpace(req.Search))
	for _, acc := range accounts {
		if globalNeedle != "" {
			haystack := strings.ToLower(strings.Join([]string{
				strconv.Itoa(acc.ID),
				getExchangeName(acc),
				acc.AccountName,
				getNote(acc),
				strconv.Itoa(acc.Priority),
				getStatus(acc),
			}, " "))
			if !strings.Contains(haystack, globalNeedle) {
				continue
			}
		}

		columnMatched := true
		for _, col := range req.Columns {
			if strings.TrimSpace(col.Search.Value) == "" {
				continue
			}
			if !matchesColumn(acc, col.Data, col.Search.Value) {
				columnMatched = false
				break
			}
		}
		if columnMatched {
			filtered = append(filtered, acc)
		}
	}

	// Сортировка по колонкам из DataTables.
	if len(req.Order) > 0 {
		lessForColumn := func(a, b *models.ExchangeAccount, colIdx int, dir string) bool {
			asc := strings.ToLower(dir) != "desc"
			compareStrings := func(x, y string) bool {
				if asc {
					return strings.ToLower(x) < strings.ToLower(y)
				}
				return strings.ToLower(x) > strings.ToLower(y)
			}

			switch colIdx {
			case 1: // id
				if asc {
					return a.ID < b.ID
				}
				return a.ID > b.ID
			case 2: // exchange_name (JOIN)
				return compareStrings(getExchangeName(a), getExchangeName(b))
			case 3: // account_name
				return compareStrings(a.AccountName, b.AccountName)
			case 4: // note
				return compareStrings(getNote(a), getNote(b))
			case 5: // priority
				if asc {
					return a.Priority < b.Priority
				}
				return a.Priority > b.Priority
			case 6: // status
				return compareStrings(getStatus(a), getStatus(b))
			default:
				if asc {
					return a.ID < b.ID
				}
				return a.ID > b.ID
			}
		}

		// DataTables поддерживает multi-sort: применяем от последнего к первому.
		for i := len(req.Order) - 1; i >= 0; i-- {
			ord := req.Order[i]
			sort.SliceStable(filtered, func(a, b int) bool {
				return lessForColumn(filtered[a], filtered[b], ord.Column, ord.Dir)
			})
		}
	}

	recordsFiltered := len(filtered)

	// Пагинация вручную
	start := req.Start
	if start < 0 {
		start = 0
	}
	end := start + req.Length
	if req.Length <= 0 {
		end = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	if start > end {
		start = end
	}
	page := filtered[start:end]

	aaData := make([]map[string]interface{}, len(page))
	for i, acc := range page {
		status := getStatus(acc)
		exchangeName := getExchangeName(acc)
		note := getNote(acc)
		aaData[i] = map[string]interface{}{
			"chbx":          "",
			"DT_RowId":      "row_" + strconv.Itoa(acc.ID),
			"id":            acc.ID,
			"exchange_name": exchangeName,
			"exchange_id":   acc.ExID,
			"account_name":  acc.AccountName,
			"priority":      acc.Priority,
			"status":        status,
			"note":          note,
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
	if acc.Active && acc.ExchangeActive {
		status = "Active"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              acc.ID,
		"exchange_id":     acc.ExID,
		"exchange_active": acc.ExchangeActive,
		"account_name":    acc.AccountName,
		"priority":        acc.Priority,
		"status":          status,
		"has_api_key":     strings.TrimSpace(acc.ApiKey) != "",
		"has_secret_key":  strings.TrimSpace(acc.SecretKey) != "",
		"has_add_key":     acc.AddKey != nil && strings.TrimSpace(*acc.AddKey) != "",
		"note":            acc.Note,
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
	exchange, err := eac.service.ExchangeRepo().FindByID(exid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exchange not found"})
		return
	}
	if !exchange.Active {
		status = "Blocked"
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
	exchange, err := eac.service.ExchangeRepo().FindByID(exid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exchange not found"})
		return
	}
	if !exchange.Active {
		status = "Blocked"
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
