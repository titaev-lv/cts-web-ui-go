package services

import (
	"fmt"
	"strings"
	"time"
)

type CSVImportRequest struct {
	PositionID int
	ExchangeID int
	Contract   string
	StartUTC   *time.Time
	StopUTC    *time.Time
	Content    []byte
}

type CSVImporter interface {
	Import(req CSVImportRequest) (int, error)
}

func (s *PositionService) getCSVImporter(exchangeID int) (CSVImporter, error) {
	switch exchangeID {
	case 7:
		return &BybitCSVImporter{repo: s.repo}, nil
	default:
		return nil, fmt.Errorf("csv import is not configured for exchange id=%d", exchangeID)
	}
}

func normalizeCSVDecimal(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" || value == "--" || value == "-" {
		return "0", nil
	}
	value = strings.ReplaceAll(value, ",", "")
	value = strings.TrimPrefix(value, "+")
	if value == "" || value == "-" {
		return "0", nil
	}
	return value, nil
}

func isZeroDecimal(raw string) bool {
	value := strings.TrimSpace(raw)
	if value == "" {
		return true
	}
	value = strings.TrimPrefix(value, "+")
	value = strings.TrimPrefix(value, "-")
	value = strings.TrimLeft(value, "0")
	value = strings.TrimPrefix(value, ".")
	value = strings.ReplaceAll(value, ".", "")
	value = strings.TrimLeft(value, "0")
	return value == ""
}

func absDecimalString(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "+")
	value = strings.TrimPrefix(value, "-")
	if value == "" {
		return "0"
	}
	return value
}

func parseCSVDateUTC(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	layouts := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.UTC)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
