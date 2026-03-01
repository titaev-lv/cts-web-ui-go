package services

import (
	"ctweb/internal/repositories"
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
	"time"
)

const dateTimeFormat = "2006-01-02 15:04:05"

type PositionService struct {
	repo *repositories.PositionRepository
}

func NewPositionService() *PositionService {
	return &PositionService{repo: repositories.NewPositionRepository()}
}

func (s *PositionService) parseDateTimeInUserTZ(value, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	dt, err := time.ParseInLocation(dateTimeFormat, strings.TrimSpace(value), loc)
	if err != nil {
		return time.Time{}, err
	}
	return dt.UTC(), nil
}

func (s *PositionService) normalizeMarket(market string) string {
	if strings.EqualFold(strings.TrimSpace(market), "spot") {
		return "SPOT"
	}
	return "FUTURES"
}

func (s *PositionService) GetPositionsData(userID, start, length int) (int, []map[string]interface{}, error) {
	count, err := s.repo.CountPositionsByUser(userID)
	if err != nil {
		return 0, nil, err
	}

	data, err := s.repo.GetPositions(userID, length, start)
	if err != nil {
		return 0, nil, err
	}

	rows := make([]map[string]interface{}, 0, len(data))
	for _, item := range data {
		row := map[string]interface{}{
			"POSITION_ID":        item.PositionID,
			"CONTRACT_NAME":      html.EscapeString(item.ContractName),
			"EXCHANGE_NAME":      html.EscapeString(item.ExchangeName),
			"MARKET_TYPE":        html.EscapeString(item.MarketType),
			"STATUS":             item.Status,
			"FINAL_POSITION":     nil,
			"FINAL_AVG_PRICE":    nil,
			"FEE_BASE_TOTAL":     nil,
			"FEE_TOTAL":          nil,
			"FUNDING_TOTAL":      nil,
			"TOTAL_REALIZED_PNL": nil,
		}
		if item.FinalPosition != nil {
			row["FINAL_POSITION"] = *item.FinalPosition
		}
		if item.FinalAvgPrice != nil {
			row["FINAL_AVG_PRICE"] = *item.FinalAvgPrice
		}
		if item.FeeBaseTotal != nil {
			row["FEE_BASE_TOTAL"] = *item.FeeBaseTotal
		}
		if item.FeeTotal != nil {
			row["FEE_TOTAL"] = *item.FeeTotal
		}
		if item.FundingTotal != nil {
			row["FUNDING_TOTAL"] = *item.FundingTotal
		}
		if item.TotalRealizedPnL != nil {
			row["TOTAL_REALIZED_PNL"] = *item.TotalRealizedPnL
		}
		rows = append(rows, row)
	}

	return count, rows, nil
}

func (s *PositionService) CreatePosition(userID int, userTimezone, name string, exchangeID int, startDate, market string) (bool, string) {
	if strings.TrimSpace(name) == "" {
		return false, `Filed "Contract Name" is empty`
	}
	if exchangeID <= 0 {
		return false, `Filed "Exchange" is empty`
	}
	if strings.TrimSpace(startDate) == "" {
		return false, `Filed "Sart Date" is empty`
	}
	if strings.TrimSpace(market) == "" {
		return false, `Filed "Market" is empty`
	}

	startUTC, err := s.parseDateTimeInUserTZ(startDate, userTimezone)
	if err != nil {
		return false, "Error format and create Position Date"
	}

	if _, err := time.Parse(dateTimeFormat, strings.TrimSpace(startDate)); err != nil {
		return false, "Error format Start Date"
	}

	if err := s.repo.CreatePosition(strings.TrimSpace(name), exchangeID, startUTC, s.normalizeMarket(market), userID); err != nil {
		return false, "Erorr create position"
	}

	return true, ""
}

func (s *PositionService) EditPosition(userID int, userTimezone string, positionID int, name string, exchangeID int, startDate string) (bool, string) {
	if positionID <= 0 {
		return false, "Failed Position ID"
	}
	if strings.TrimSpace(name) == "" {
		return false, `Filed "Contract Name" is empty`
	}
	if exchangeID <= 0 {
		return false, `Filed "Exchange" is empty`
	}
	if strings.TrimSpace(startDate) == "" {
		return false, `Filed "Sart Date" is empty`
	}

	startUTC, err := s.parseDateTimeInUserTZ(startDate, userTimezone)
	if err != nil {
		return false, "Error format and create Position Date"
	}

	if _, err := time.Parse(dateTimeFormat, strings.TrimSpace(startDate)); err != nil {
		return false, "Error format Start Date"
	}

	updated, err := s.repo.EditPosition(positionID, userID, strings.TrimSpace(name), exchangeID, startUTC)
	if err != nil {
		return false, "Error edit position"
	}
	if !updated {
		return false, "Failed Position ID"
	}

	return true, ""
}

func (s *PositionService) GetPosition(userID int, userTimezone string, positionID int) (map[string]interface{}, bool, string) {
	item, err := s.repo.GetPositionByID(userID, positionID)
	if err != nil {
		return nil, false, "Empty Position Data"
	}
	if item == nil {
		return nil, false, "Empty Position Data"
	}

	loc, tzErr := time.LoadLocation(userTimezone)
	if tzErr != nil {
		loc = time.UTC
	}

	opened := ""
	if item.Created != nil {
		opened = item.Created.In(loc).Format(dateTimeFormat)
	}
	closed := "—"
	if item.Closed != nil {
		closed = item.Closed.In(loc).Format(dateTimeFormat)
	}

	amount := ""
	if item.FinalPosition != nil {
		amount = strconv.FormatFloat(*item.FinalPosition, 'f', -1, 64)
	}
	avg := ""
	if item.FinalAvgPrice != nil {
		avg = strconv.FormatFloat(*item.FinalAvgPrice, 'f', -1, 64)
	}
	feeBase := ""
	if item.FeeBaseTotal != nil {
		feeBase = strconv.FormatFloat(*item.FeeBaseTotal, 'f', -1, 64)
	}
	fee := ""
	if item.FeeTotal != nil {
		fee = strconv.FormatFloat(*item.FeeTotal, 'f', -1, 64)
	}
	funding := ""
	if item.FundingTotal != nil {
		funding = strconv.FormatFloat(*item.FundingTotal, 'f', -1, 64)
	}
	realized := ""
	if item.TotalRealizedPnL != nil {
		realized = strconv.FormatFloat(*item.TotalRealizedPnL, 'f', -1, 64)
	}

	result := map[string]interface{}{
		"POSITION_ID":        item.PositionID,
		"CONTRACT_NAME":      html.EscapeString(item.ContractName),
		"EXCHANGE_NAME":      html.EscapeString(item.ExchangeName),
		"MARKET_TYPE":        html.EscapeString(item.MarketType),
		"STATUS":             html.EscapeString(strings.ToUpper(item.Status)),
		"OPENED":             opened,
		"CLOSED":             closed,
		"AMOUNT":             amount,
		"AVG_PRICE":          avg,
		"FEE_BASE_CURR":      feeBase,
		"FEE_QUOTE_CURR":     fee,
		"FUNDING":            funding,
		"TOTAL_REALIZED_PNL": realized,
		"TRANS_COUNT":        strconv.Itoa(item.TransCount),
	}

	return result, true, ""
}

func (s *PositionService) GetTransactions(userID int, userTimezone string, positionID, start, length int) (int, []map[string]interface{}, string) {
	count, err := s.repo.CountTransactionsByPosition(positionID, userID)
	if err != nil {
		return 0, nil, ""
	}

	items, err := s.repo.GetTransactionsByPosition(positionID, userID, length, start)
	if err != nil {
		return 0, nil, ""
	}

	loc, tzErr := time.LoadLocation(userTimezone)
	if tzErr != nil {
		loc = time.UTC
	}

	rows := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		transDate := ""
		if item.TransDate != nil {
			transDate = item.TransDate.In(loc).Format(dateTimeFormat)
		}

		rows = append(rows, map[string]interface{}{
			"ID":         item.ID,
			"TYPE":       item.Type,
			"PRICE":      item.Price,
			"VOLUME":     item.Volume,
			"FEE_BASE":   item.FeeBase,
			"FEE":        item.Fee,
			"FUNDING":    item.Funding,
			"TRANS_DATE": transDate,
		})
	}

	return count, rows, ""
}

func (s *PositionService) CreateTransaction(userID int, userTimezone string, req map[string]string) (bool, string) {
	positionID, _ := strconv.Atoi(req["add_trans_position"])
	if positionID <= 0 {
		positionID, _ = strconv.Atoi(req["position_id"])
	}
	typeValue := strings.TrimSpace(req["add_trans_type"])
	transDate := strings.TrimSpace(req["add_trans_date"])
	action := strings.TrimSpace(req["add_trans_action"])

	if positionID <= 0 || typeValue == "" {
		return false, "Position data ERROR"
	}

	marketType, err := s.repo.GetPositionMarketType(positionID, userID)
	if err != nil {
		return false, "Position data ERROR"
	}
	if marketType != "FUTURES" && marketType != "SPOT" {
		return false, "Market type ERROR"
	}

	parsedDateUTC, err := s.parseDateTimeInUserTZ(transDate, userTimezone)
	if err != nil {
		return false, "Error format and create Transaction Date"
	}

	parseNum := func(name string) float64 {
		v := strings.TrimSpace(req[name])
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	funding := parseNum("add_trans_funding")
	volume := parseNum("add_trans_volume")
	price := parseNum("add_trans_price")
	fee := parseNum("add_trans_fee_quote")
	feeBase := parseNum("add_trans_fee_base")

	if transDate == "" {
		return false, `Filed "Transaction Date" is empty`
	}

	if marketType == "FUTURES" {
		if typeValue == "funding" {
			if req["add_trans_funding"] == "" {
				return false, `Filed "Funding" is empty`
			}
			if err := s.repo.InsertFundingTransaction(positionID, funding, parsedDateUTC); err != nil {
				return false, fmt.Sprintf("Error insert transaction %v", err)
			}
			return true, ""
		}

		if action == "" {
			return false, `Filed "Action" is empty`
		}
		if req["add_trans_price"] == "" {
			return false, `Filed "Price" is empty`
		}
		if req["add_trans_volume"] == "" {
			return false, `Filed "Volume" is empty`
		}
		if req["add_trans_fee_quote"] == "" {
			return false, `Filed "Fee" is empty`
		}

		if action == "sell" && volume > 0 {
			volume = -volume
		}
		if action == "buy" && volume < 0 {
			volume = -volume
		}
		if price < 0 {
			price = -price
		}
		if fee < 0 {
			fee = -fee
		}

		if err := s.repo.InsertTradeTransaction(positionID, price, volume, fee, 0, parsedDateUTC); err != nil {
			return false, fmt.Sprintf("Error insert transaction %v", err)
		}
		return true, ""
	}

	if action == "" {
		return false, "Action is empty"
	}
	if req["add_trans_price"] == "" {
		return false, `Filed "Price" is empty`
	}
	if req["add_trans_volume"] == "" {
		return false, `Filed "Volume" is empty`
	}
	if action == "buy" && req["add_trans_fee_base"] == "" {
		return false, `Filed "Fee" is empty`
	}
	if action != "buy" && req["add_trans_fee_quote"] == "" {
		return false, `Filed "Fee" is empty`
	}

	if action == "sell" && volume > 0 {
		volume = -volume
	}
	if action == "buy" && volume < 0 {
		volume = -volume
	}
	if price < 0 {
		price = -price
	}
	if fee < 0 {
		fee = -fee
	}
	if feeBase < 0 {
		feeBase = -feeBase
	}

	if action == "buy" {
		if err := s.repo.InsertTradeTransaction(positionID, price, volume, 0, feeBase, parsedDateUTC); err != nil {
			return false, fmt.Sprintf("Error insert transaction %v", err)
		}
		return true, ""
	}

	if err := s.repo.InsertTradeTransaction(positionID, price, volume, fee, 0, parsedDateUTC); err != nil {
		return false, fmt.Sprintf("Error insert transaction %v", err)
	}
	return true, ""
}

func (s *PositionService) GetTransactionByID(userID int, userTimezone string, positionID, transactionID int) (map[string]interface{}, bool, string) {
	if positionID <= 0 || transactionID <= 0 {
		return nil, false, "Empty ID"
	}

	item, err := s.repo.GetTransactionByID(userID, positionID, transactionID)
	if err != nil {
		return nil, false, "Empty transaction data"
	}
	if item == nil {
		return nil, false, "Empty transaction data"
	}

	loc, tzErr := time.LoadLocation(userTimezone)
	if tzErr != nil {
		loc = time.UTC
	}

	transDate := ""
	if item.TransDate != nil {
		transDate = item.TransDate.In(loc).Format(dateTimeFormat)
	}

	result := map[string]interface{}{
		"ID":         item.ID,
		"TYPE":       item.Type,
		"PRICE":      item.Price,
		"VOLUME":     item.Volume,
		"FEE_BASE":   item.FeeBase,
		"FEE":        item.Fee,
		"FUNDING":    item.Funding,
		"TRANS_DATE": transDate,
	}

	return result, true, ""
}

func (s *PositionService) EditTransaction(userID int, userTimezone string, req map[string]string) (bool, string) {
	transactionID, _ := strconv.Atoi(req["edit_trans_id"])
	positionID, _ := strconv.Atoi(req["edit_trans_position"])
	if positionID <= 0 {
		positionID, _ = strconv.Atoi(req["position_id"])
	}
	typeValue := strings.TrimSpace(req["edit_trans_type"])
	transDate := strings.TrimSpace(req["edit_trans_date"])
	action := strings.TrimSpace(req["edit_trans_action"])

	if transactionID <= 0 {
		return false, "Transaction ID ERROR"
	}
	if positionID <= 0 || typeValue == "" {
		return false, "Position data ERROR"
	}

	marketType, err := s.repo.GetPositionMarketType(positionID, userID)
	if err != nil {
		return false, "Position data ERROR"
	}
	if marketType != "FUTURES" && marketType != "SPOT" {
		return false, "Market type ERROR"
	}

	parsedDateUTC, err := s.parseDateTimeInUserTZ(transDate, userTimezone)
	if err != nil {
		return false, "Error format and create Transaction Date"
	}

	parseNum := func(name string) float64 {
		v := strings.TrimSpace(req[name])
		if v == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	funding := parseNum("edit_trans_funding")
	volume := parseNum("edit_trans_volume")
	price := parseNum("edit_trans_price")
	fee := parseNum("edit_trans_fee_quote")
	feeBase := parseNum("edit_trans_fee_base")

	if transDate == "" {
		return false, `Filed "Transaction Date" is empty`
	}

	if marketType == "FUTURES" {
		if typeValue == "funding" {
			if strings.TrimSpace(req["edit_trans_funding"]) == "" {
				return false, `Filed "Funding" is empty`
			}
			updated, updErr := s.repo.UpdateTransactionByID(userID, positionID, transactionID, 0, 0, 0, 0, funding, parsedDateUTC, "FUNDING")
			if updErr != nil {
				return false, fmt.Sprintf("Error edit transaction %v", updErr)
			}
			if !updated {
				return false, "Transaction not found"
			}
			return true, ""
		}

		if action == "" {
			return false, `Filed "Action" is empty`
		}
		if strings.TrimSpace(req["edit_trans_price"]) == "" {
			return false, `Filed "Price" is empty`
		}
		if strings.TrimSpace(req["edit_trans_volume"]) == "" {
			return false, `Filed "Volume" is empty`
		}
		if strings.TrimSpace(req["edit_trans_fee_quote"]) == "" {
			return false, `Filed "Fee" is empty`
		}

		if action == "sell" && volume > 0 {
			volume = -volume
		}
		if action == "buy" && volume < 0 {
			volume = -volume
		}
		if price < 0 {
			price = -price
		}
		if fee < 0 {
			fee = -fee
		}

		updated, updErr := s.repo.UpdateTransactionByID(userID, positionID, transactionID, price, volume, fee, 0, 0, parsedDateUTC, "TRADE")
		if updErr != nil {
			return false, fmt.Sprintf("Error edit transaction %v", updErr)
		}
		if !updated {
			return false, "Transaction not found"
		}
		return true, ""
	}

	if action == "" {
		return false, "Action is empty"
	}
	if strings.TrimSpace(req["edit_trans_price"]) == "" {
		return false, `Filed "Price" is empty`
	}
	if strings.TrimSpace(req["edit_trans_volume"]) == "" {
		return false, `Filed "Volume" is empty`
	}
	if action == "buy" && strings.TrimSpace(req["edit_trans_fee_base"]) == "" {
		return false, `Filed "Fee" is empty`
	}
	if action != "buy" && strings.TrimSpace(req["edit_trans_fee_quote"]) == "" {
		return false, `Filed "Fee" is empty`
	}

	if action == "sell" && volume > 0 {
		volume = -volume
	}
	if action == "buy" && volume < 0 {
		volume = -volume
	}
	if price < 0 {
		price = -price
	}
	if fee < 0 {
		fee = -fee
	}
	if feeBase < 0 {
		feeBase = -feeBase
	}

	tradeFee := fee
	tradeFeeBase := 0.0
	if action == "buy" {
		tradeFee = 0
		tradeFeeBase = feeBase
	}

	updated, updErr := s.repo.UpdateTransactionByID(userID, positionID, transactionID, price, volume, tradeFee, tradeFeeBase, 0, parsedDateUTC, "TRADE")
	if updErr != nil {
		return false, fmt.Sprintf("Error edit transaction %v", updErr)
	}
	if !updated {
		return false, "Transaction not found"
	}

	return true, ""
}

func (s *PositionService) UploadTransactionsCSV(userID int, userTimezone string, req map[string]string, content []byte) (int, bool, string) {
	positionID, _ := strconv.Atoi(strings.TrimSpace(req["import_trans_csv_position"]))
	exchangeID, _ := strconv.Atoi(strings.TrimSpace(req["import_trans_csv_exchange"]))
	contract := strings.TrimSpace(req["import_trans_csv_contract_name"])
	startDateRaw := strings.TrimSpace(req["import_trans_csv_start_date"])
	stopDateRaw := strings.TrimSpace(req["import_trans_csv_stop_date"])

	if positionID <= 0 {
		return 0, false, `Filed "Position ID" is empty`
	}
	if exchangeID <= 0 {
		return 0, false, `Filed "Exchange" is empty`
	}
	if contract == "" {
		return 0, false, `Filed "Contract" is empty`
	}
	if len(content) == 0 {
		return 0, false, "Can not read data from file"
	}

	marketType, err := s.repo.GetPositionMarketType(positionID, userID)
	if err != nil {
		return 0, false, "Position data ERROR"
	}
	if strings.TrimSpace(marketType) == "" {
		return 0, false, "Position data ERROR"
	}

	var startUTC *time.Time
	if startDateRaw != "" {
		startValue, err := s.parseDateTimeInUserTZ(startDateRaw, userTimezone)
		if err != nil {
			return 0, false, "Error format and create Start Date"
		}
		startUTC = &startValue
	}

	var stopUTC *time.Time
	if stopDateRaw != "" {
		stopValue, err := s.parseDateTimeInUserTZ(stopDateRaw, userTimezone)
		if err != nil {
			return 0, false, "Error format and create Stop Date"
		}
		if !strings.Contains(stopDateRaw, ".") {
			stopValue = stopValue.Add(999 * time.Millisecond)
		}
		stopUTC = &stopValue
	}

	importer, err := s.getCSVImporter(exchangeID)
	if err != nil {
		return 0, false, "CSV import is not configured for selected exchange"
	}

	inserted, importErr := importer.Import(CSVImportRequest{
		PositionID: positionID,
		ExchangeID: exchangeID,
		Contract:   contract,
		StartUTC:   startUTC,
		StopUTC:    stopUTC,
		Content:    content,
	})
	if importErr != nil {
		return inserted, false, importErr.Error()
	}

	return inserted, true, ""
}

func (s *PositionService) ClosePosition(userID, positionID int) (bool, string) {
	if positionID <= 0 {
		return false, "Failed Position ID"
	}

	status, amount, err := s.repo.GetPositionStatusAndAmount(positionID, userID)
	if err != nil {
		return false, "Can't close. Position not opened"
	}

	if status != "OPEN" {
		return false, "Can't close. Position not opened"
	}
	if math.Abs(amount) > 1e-16 {
		return false, "Can't close. Position not 0"
	}

	updated, err := s.repo.ClosePosition(positionID, userID)
	if err != nil {
		return false, "Can't close. Position not opened"
	}

	return updated, ""
}

func (s *PositionService) DeletePosition(userID, positionID int) (bool, string) {
	if positionID <= 0 {
		return false, "Failed Position ID"
	}

	deleted, err := s.repo.DeletePosition(positionID, userID)
	if err != nil {
		return false, "Failed Position ID"
	}

	return deleted, ""
}

func (s *PositionService) DeleteTransactions(userID, positionID int, transactionIDs []int) (bool, string) {
	if positionID <= 0 {
		return false, "Failed Position ID"
	}
	if len(transactionIDs) == 0 {
		return false, "No transactions selected"
	}

	cleanIDs := make([]int, 0, len(transactionIDs))
	for _, id := range transactionIDs {
		if id > 0 {
			cleanIDs = append(cleanIDs, id)
		}
	}
	if len(cleanIDs) == 0 {
		return false, "No transactions selected"
	}

	affected, err := s.repo.DeleteTransactionsByIDs(userID, positionID, cleanIDs)
	if err != nil {
		return false, "Failed delete transactions"
	}
	if affected == 0 {
		return false, "No transactions deleted"
	}

	return true, ""
}

func (s *PositionService) Repo() *repositories.PositionRepository {
	return s.repo
}
