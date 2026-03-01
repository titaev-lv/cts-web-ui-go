package controllers

import (
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"ctweb/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PositionController struct {
	service *services.PositionService
}

func NewPositionController() *PositionController {
	return &PositionController{service: services.NewPositionService()}
}

func (pc *PositionController) List(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	exchanges, _ := repositories.NewExchangeRepository().FindAll()
	sort.Slice(exchanges, func(i, j int) bool {
		return exchanges[i].Name < exchanges[j].Name
	})

	nowMoscow := time.Now().In(time.FixedZone("MSK", 3*60*60)).Format("2006-01-02 15:04:05")

	c.HTML(http.StatusOK, "positions/index.html", gin.H{
		"Title":     "Trade Positions",
		"User":      user.(*models.User),
		"Exchanges": exchanges,
		"Now":       nowMoscow,
	})
}

func (pc *PositionController) PositionPage(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	exchanges, _ := repositories.NewExchangeRepository().FindAll()
	sort.Slice(exchanges, func(i, j int) bool {
		return exchanges[i].Name < exchanges[j].Name
	})

	exchangeImportCSV := make([]*models.Exchange, 0)
	for _, ex := range exchanges {
		if ex.ID == 7 {
			exchangeImportCSV = append(exchangeImportCSV, ex)
		}
	}

	nowMoscow := time.Now().In(time.FixedZone("MSK", 3*60*60)).Format("2006-01-02 15:04:05")

	c.HTML(http.StatusOK, "positions/position.html", gin.H{
		"Title":             "Position",
		"User":              user.(*models.User),
		"Exchanges":         exchanges,
		"ExchangeImportCSV": exchangeImportCSV,
		"Now":               nowMoscow,
	})
}

func (pc *PositionController) AjaxGetPositions(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	start, _ := strconv.Atoi(c.DefaultPostForm("start", "0"))
	length, _ := strconv.Atoi(c.DefaultPostForm("length", "50"))
	if length <= 0 {
		length = 50
	}

	count, rows, err := pc.service.GetPositionsData(user.ID, start, length)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"recordsTotal":    0,
			"recordsFiltered": 0,
			"aaData":          []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordsTotal":    count,
		"recordsFiltered": count,
		"aaData":          rows,
	})
}

func (pc *PositionController) AjaxCreatePosition(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	name := c.PostForm("add_position_name_contract")
	exchangeID, _ := strconv.Atoi(c.PostForm("add_position_exchange"))
	startDate := c.PostForm("add_position_date_start")
	market := c.PostForm("add_position_market")

	success, errText := pc.service.CreatePosition(user.ID, user.Timezone, name, exchangeID, startDate, market)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxEditPosition(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	positionID, _ := strconv.Atoi(c.PostForm("position_id"))
	name := c.PostForm("name_contract")
	exchangeID, _ := strconv.Atoi(c.PostForm("exchange_id"))
	startDate := c.PostForm("date_start")

	success, errText := pc.service.EditPosition(user.ID, user.Timezone, positionID, name, exchangeID, startDate)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxGetPosition(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	positionID, _ := strconv.Atoi(c.PostForm("position_id"))
	if positionID <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "Empty ID"})
		return
	}

	row, success, errText := pc.service.GetPosition(user.ID, user.Timezone, positionID)
	if !success {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": errText})
		return
	}

	row["success"] = true
	row["error"] = false
	c.JSON(http.StatusOK, row)
}

func (pc *PositionController) AjaxGetTransactions(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	transID, _ := strconv.Atoi(c.DefaultPostForm("trans_id", "0"))
	if transID > 0 {
		positionID, _ := strconv.Atoi(c.DefaultPostForm("position_id", "0"))
		row, success, errText := pc.service.GetTransactionByID(user.ID, user.Timezone, positionID, transID)
		if !success {
			c.JSON(http.StatusOK, gin.H{"success": false, "error": errText})
			return
		}

		row["success"] = true
		row["error"] = false
		c.JSON(http.StatusOK, row)
		return
	}

	positionID, _ := strconv.Atoi(c.DefaultPostForm("position_id", "0"))
	start, _ := strconv.Atoi(c.DefaultPostForm("start", "0"))
	length, _ := strconv.Atoi(c.DefaultPostForm("length", "50"))
	if length <= 0 {
		length = 50
	}

	if positionID <= 0 {
		c.JSON(http.StatusOK, gin.H{"recordsTotal": 0, "recordsFiltered": 0, "aaData": []interface{}{}})
		return
	}

	count, rows, _ := pc.service.GetTransactions(user.ID, user.Timezone, positionID, start, length)
	c.JSON(http.StatusOK, gin.H{
		"recordsTotal":    count,
		"recordsFiltered": count,
		"aaData":          rows,
	})
}

func (pc *PositionController) AjaxCreateTransaction(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	req := map[string]string{
		"add_trans_position":  c.PostForm("add_trans_position"),
		"position_id":         c.PostForm("position_id"),
		"add_trans_type":      c.PostForm("add_trans_type"),
		"add_trans_date":      c.PostForm("add_trans_date"),
		"add_trans_funding":   c.PostForm("add_trans_funding"),
		"add_trans_volume":    c.PostForm("add_trans_volume"),
		"add_trans_action":    c.PostForm("add_trans_action"),
		"add_trans_price":     c.PostForm("add_trans_price"),
		"add_trans_fee_quote": c.PostForm("add_trans_fee_quote"),
		"add_trans_fee_base":  c.PostForm("add_trans_fee_base"),
	}

	success, errText := pc.service.CreateTransaction(user.ID, user.Timezone, req)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxEditTransaction(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	req := map[string]string{
		"edit_trans_id":        c.PostForm("edit_trans_id"),
		"edit_trans_position":  c.PostForm("edit_trans_position"),
		"position_id":          c.PostForm("position_id"),
		"edit_trans_type":      c.PostForm("edit_trans_type"),
		"edit_trans_date":      c.PostForm("edit_trans_date"),
		"edit_trans_funding":   c.PostForm("edit_trans_funding"),
		"edit_trans_volume":    c.PostForm("edit_trans_volume"),
		"edit_trans_action":    c.PostForm("edit_trans_action"),
		"edit_trans_price":     c.PostForm("edit_trans_price"),
		"edit_trans_fee_quote": c.PostForm("edit_trans_fee_quote"),
		"edit_trans_fee_base":  c.PostForm("edit_trans_fee_base"),
	}

	success, errText := pc.service.EditTransaction(user.ID, user.Timezone, req)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxUploadTransactionCSV(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	req := map[string]string{
		"import_trans_csv_position":      c.PostForm("import_trans_csv_position"),
		"import_trans_csv_exchange":      c.PostForm("import_trans_csv_exchange"),
		"import_trans_csv_start_date":    c.PostForm("import_trans_csv_start_date"),
		"import_trans_csv_stop_date":     c.PostForm("import_trans_csv_stop_date"),
		"import_trans_csv_contract_name": c.PostForm("import_trans_csv_contract_name"),
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fileHeader, err = c.FormFile("import_trans_csv_file")
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "File Not Attached",
			"success": false,
			"data":    false,
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Can not read data from file",
			"success": false,
			"data":    false,
		})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":   "Can not read data from file",
			"success": false,
			"data":    false,
		})
		return
	}

	inserted, success, errText := pc.service.UploadTransactionsCSV(user.ID, user.Timezone, req, content)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
		"data":    inserted,
	})
}

func (pc *PositionController) AjaxDeleteTransaction(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	positionID, _ := strconv.Atoi(c.PostForm("position_id"))
	rawIDs := strings.TrimSpace(c.PostForm("transaction_ids"))
	if rawIDs == "" {
		c.JSON(http.StatusOK, gin.H{"error": "No transactions selected", "success": false})
		return
	}

	var ids []int
	if err := json.Unmarshal([]byte(rawIDs), &ids); err != nil {
		var strIDs []string
		if err2 := json.Unmarshal([]byte(rawIDs), &strIDs); err2 != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Invalid transactions payload", "success": false})
			return
		}
		for _, value := range strIDs {
			id, convErr := strconv.Atoi(strings.TrimSpace(value))
			if convErr == nil {
				ids = append(ids, id)
			}
		}
	}

	success, errText := pc.service.DeleteTransactions(user.ID, positionID, ids)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxClosePosition(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	positionID, _ := strconv.Atoi(c.PostForm("position_id"))
	success, errText := pc.service.ClosePosition(user.ID, positionID)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxDeletePosition(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user := userVal.(*models.User)

	positionID, _ := strconv.Atoi(c.PostForm("position_id"))
	success, errText := pc.service.DeletePosition(user.ID, positionID)
	c.JSON(http.StatusOK, gin.H{
		"error":   boolOrError(errText),
		"success": success,
	})
}

func (pc *PositionController) AjaxKucoinPrice(c *gin.Context) {
	if _, exists := c.Get("user"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	symbol := strings.TrimSpace(c.PostForm("symbol"))
	market := strings.ToUpper(strings.TrimSpace(c.DefaultPostForm("market", "SPOT")))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	var endpoint string
	if market == "SPOT" {
		mapped := strings.ReplaceAll(symbol, "/", "-")
		endpoint = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=" + url.QueryEscape(mapped)
	} else {
		mapped := strings.ReplaceAll(symbol, "/", "") + "M"
		endpoint = "https://api-futures.kucoin.com/api/v1/market/ticker?symbol=" + url.QueryEscape(mapped)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Failed to get KuCoin price"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{"error": fmt.Sprintf("HTTP %d", resp.StatusCode)})
		return
	}

	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.Data == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid response format"})
		return
	}

	price := ""
	if market == "SPOT" {
		if value, ok := payload.Data["price"].(string); ok {
			price = value
		}
	} else {
		if value, ok := payload.Data["last"].(string); ok {
			price = value
		}
	}

	if strings.TrimSpace(price) == "" {
		c.JSON(http.StatusOK, gin.H{"error": "Price not found in response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "price": price})
}

func (pc *PositionController) AjaxKucoinToken(c *gin.Context) {
	if _, exists := c.Get("user"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	market := strings.ToUpper(strings.TrimSpace(c.DefaultPostForm("market", "SPOT")))
	endpoint := "https://api.kucoin.com/api/v1/bullet-public"
	if market != "SPOT" {
		endpoint = "https://api-futures.kucoin.com/api/v1/bullet-public"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Failed to build request"})
		return
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Failed to get KuCoin token"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{"error": fmt.Sprintf("HTTP %d", resp.StatusCode)})
		return
	}

	var payload struct {
		Data any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || payload.Data == nil {
		c.JSON(http.StatusOK, gin.H{"error": "Invalid response format"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": payload.Data})
}

func boolOrError(errText string) interface{} {
	if errText == "" {
		return false
	}
	return errText
}
