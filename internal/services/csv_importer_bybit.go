package services

import (
	"bytes"
	"ctweb/internal/repositories"
	"encoding/csv"
	"fmt"
	"strings"
)

type BybitCSVImporter struct {
	repo *repositories.PositionRepository
}

func (i *BybitCSVImporter) Import(req CSVImportRequest) (int, error) {
	reader := csv.NewReader(bytes.NewReader(req.Content))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("Error parse file")
	}
	if len(records) < 2 {
		return 0, fmt.Errorf("Empty DATA")
	}

	cleanHeader := func(value string) string {
		cleaned := strings.TrimSpace(value)
		cleaned = strings.TrimPrefix(cleaned, "\uFEFF")
		return cleaned
	}

	header := records[0]
	for index := range header {
		header[index] = cleanHeader(header[index])
	}

	findIndex := func(columnName string) int {
		for index, name := range header {
			if name == columnName {
				return index
			}
		}
		return -1
	}

	posContract := findIndex("Contract")
	posTransDate := findIndex("Time")
	posType := findIndex("Type")
	posDirection := findIndex("Direction")
	posQuantity := findIndex("Quantity")
	posPrice := findIndex("Filled Price")
	posFunding := findIndex("Funding")
	posFee := findIndex("Fee Paid")
	posOrderID := findIndex("OrderId")
	posTradeID := findIndex("TradeId")

	if posContract < 0 || posTransDate < 0 || posType < 0 || posDirection < 0 || posQuantity < 0 || posPrice < 0 || posFunding < 0 || posFee < 0 {
		return 0, fmt.Errorf("Error parse file")
	}

	inserted := 0
	matched := 0

	for rowIndex := len(records) - 1; rowIndex >= 1; rowIndex-- {
		row := records[rowIndex]
		if len(row) <= 1 {
			continue
		}

		maxIndex := posFee
		for _, index := range []int{posContract, posTransDate, posType, posDirection, posQuantity, posPrice, posFunding, posFee, posOrderID, posTradeID} {
			if index < 0 {
				continue
			}
			if index > maxIndex {
				maxIndex = index
			}
		}
		if len(row) <= maxIndex {
			continue
		}

		transDate, parseErr := parseCSVDateUTC(row[posTransDate])
		if parseErr != nil {
			return inserted, fmt.Errorf("Error transaction date in file")
		}

		if req.Contract != strings.TrimSpace(row[posContract]) {
			continue
		}
		if req.StartUTC != nil && transDate.Before(*req.StartUTC) {
			continue
		}
		if req.StopUTC != nil && transDate.After(*req.StopUTC) {
			continue
		}
		matched++

		typeValue := strings.ToUpper(strings.TrimSpace(row[posType]))
		direction := strings.ToUpper(strings.TrimSpace(row[posDirection]))

		var sourceOrderID *string
		if posOrderID >= 0 {
			value := strings.TrimSpace(row[posOrderID])
			if value != "" {
				sourceOrderID = &value
			}
		}

		var sourceTradeID *string
		if posTradeID >= 0 {
			value := strings.TrimSpace(row[posTradeID])
			if value != "" {
				sourceTradeID = &value
			}
		}

		if typeValue == "SETTLEMENT" {
			funding, normErr := normalizeCSVDecimal(row[posFunding])
			if normErr != nil {
				return inserted, fmt.Errorf("Error parse file")
			}
			insertedNow, err := i.repo.InsertFundingTransactionImport(req.PositionID, funding, transDate, sourceOrderID, sourceTradeID)
			if err != nil {
				return inserted, fmt.Errorf("Error insert into DB: %v", err)
			}
			if insertedNow {
				inserted++
			}
			continue
		}

		quantity, normErr := normalizeCSVDecimal(row[posQuantity])
		if normErr != nil {
			return inserted, fmt.Errorf("Error parse file")
		}
		price, normErr := normalizeCSVDecimal(row[posPrice])
		if normErr != nil {
			return inserted, fmt.Errorf("Error parse file")
		}
		feePaid, normErr := normalizeCSVDecimal(row[posFee])
		if normErr != nil {
			return inserted, fmt.Errorf("Error parse file")
		}

		volume := absDecimalString(quantity)
		if direction != "BUY" && !isZeroDecimal(volume) {
			volume = "-" + volume
		}

		insertedNow, err := i.repo.InsertTradeTransactionImport(req.PositionID, absDecimalString(price), volume, absDecimalString(feePaid), "0", transDate, sourceOrderID, sourceTradeID)
		if err != nil {
			return inserted, fmt.Errorf("Error insert into DB: %v", err)
		}
		if insertedNow {
			inserted++
		}
	}

	if matched == 0 {
		return 0, fmt.Errorf("There is no data with the specified parameters")
	}

	return inserted, nil
}
