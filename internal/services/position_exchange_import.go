package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	exchangeImportPreviewTTL      = 15 * time.Minute
	bybitAPIBaseURL               = "https://api.bybit.com"
	bybitRecvWindow               = "10000"
	bybitHistoryMaxPages          = 10
	bybitHistoryPageSize          = 100
	exchangeImportHTTPCallTimeout = 15 * time.Second
)

var errExchangeHistoryProviderNotImplemented = errors.New("service is not implemented")

type ExchangeImportCredentials struct {
	APIKey    string
	SecretKey string
	AddKey    string
}

type ExchangeImportAccountSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExchangeImportPrepareResult struct {
	Mode                string                         `json:"mode"`
	Warning             string                         `json:"warning,omitempty"`
	Accounts            []ExchangeImportAccountSummary `json:"accounts,omitempty"`
	AutoSelectAccountID int                            `json:"auto_select_account_id,omitempty"`
}

type ExchangeImportPreviewRow struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Price            string `json:"price"`
	Volume           string `json:"volume"`
	FeeBaseCurrency  string `json:"fee_base_currency"`
	FeeQuoteCurrency string `json:"fee_quote_currency"`
	Funding          string `json:"funding"`
	TransactionDate  string `json:"transaction_date"`
}

type ExchangeImportPreviewResult struct {
	PreviewToken string                     `json:"preview_token"`
	Rows         []ExchangeImportPreviewRow `json:"rows"`
	Warning      string                     `json:"warning,omitempty"`
}

type ExchangeImportRowsResult struct {
	TotalSelected int `json:"total_selected"`
	Inserted      int `json:"inserted"`
	Skipped       int `json:"skipped"`
}

type FetchHistoryRequest struct {
	ExchangeName string
	MarketType   string
	Contract     string
	StartUTC     *time.Time
	EndUTC       *time.Time
	Credentials  ExchangeImportCredentials
}

type exchangeHistoryCandidate struct {
	Type          string
	Price         string
	Volume        string
	FeeBase       string
	FeeQuote      string
	Funding       string
	TransDateUTC  time.Time
	SourceOrderID *string
	SourceTradeID *string
}

type ExchangeHistoryProvider interface {
	Fetch(ctx context.Context, req FetchHistoryRequest) ([]exchangeHistoryCandidate, error)
}

type UnsupportedHistoryProvider struct{}

func (p *UnsupportedHistoryProvider) Fetch(_ context.Context, _ FetchHistoryRequest) ([]exchangeHistoryCandidate, error) {
	return nil, errExchangeHistoryProviderNotImplemented
}

type BybitHistoryProvider struct {
	httpClient *http.Client
}

type bybitExecutionResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		List []struct {
			Symbol    string `json:"symbol"`
			OrderID   string `json:"orderId"`
			ExecID    string `json:"execId"`
			ExecType  string `json:"execType"`
			Side      string `json:"side"`
			ExecPrice string `json:"execPrice"`
			ExecQty   string `json:"execQty"`
			ExecFee   string `json:"execFee"`
			ExecValue string `json:"execValue"`
			ExecTime  string `json:"execTime"`
		} `json:"list"`
		NextPageCursor string `json:"nextPageCursor"`
	} `json:"result"`
}

type exchangeImportPreviewEntry struct {
	Token      string
	UserID     int
	PositionID int
	Rows       []exchangeHistoryCandidate
	ExpiresAt  time.Time
}

type exchangeImportPreviewStore struct {
	mu      sync.Mutex
	entries map[string]*exchangeImportPreviewEntry
}

func newExchangeImportPreviewStore() *exchangeImportPreviewStore {
	return &exchangeImportPreviewStore{
		entries: make(map[string]*exchangeImportPreviewEntry),
	}
}

func (s *exchangeImportPreviewStore) save(userID, positionID int, rows []exchangeHistoryCandidate) (string, error) {
	token, err := generateExchangePreviewToken()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	entry := &exchangeImportPreviewEntry{
		Token:      token,
		UserID:     userID,
		PositionID: positionID,
		Rows:       rows,
		ExpiresAt:  now.Add(exchangeImportPreviewTTL),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked(now)
	s.entries[token] = entry
	return token, nil
}

func (s *exchangeImportPreviewStore) load(token string) (*exchangeImportPreviewEntry, bool) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked(now)
	entry, ok := s.entries[token]
	if !ok {
		return nil, false
	}
	return entry, true
}

func (s *exchangeImportPreviewStore) delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, token)
}

func (s *exchangeImportPreviewStore) cleanupExpiredLocked(now time.Time) {
	for key, value := range s.entries {
		if value.ExpiresAt.Before(now) {
			delete(s.entries, key)
		}
	}
}

var defaultExchangeImportPreviewStore = newExchangeImportPreviewStore()

func generateExchangePreviewToken() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate preview token: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func (s *PositionService) getHistoryProvider(exchangeName string) ExchangeHistoryProvider {
	if strings.EqualFold(strings.TrimSpace(exchangeName), "Bybit") {
		return &BybitHistoryProvider{httpClient: &http.Client{Timeout: exchangeImportHTTPCallTimeout}}
	}
	return &UnsupportedHistoryProvider{}
}

func (s *PositionService) PrepareExchangeImport(userID, positionID int) (*ExchangeImportPrepareResult, error) {
	if positionID <= 0 {
		return nil, fmt.Errorf("position id is invalid")
	}

	position, err := s.repo.GetPositionExchangeAndContract(userID, positionID)
	if err != nil {
		return nil, err
	}
	if position == nil {
		return nil, fmt.Errorf("position data is empty")
	}

	accountRepo := repositories.NewExchangeAccountRepository()
	accounts, err := accountRepo.FindActiveByUserExchange(userID, position.ExchangeID)
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return &ExchangeImportPrepareResult{
			Mode:    "no_accounts",
			Warning: "No active exchange account found for this position exchange",
		}, nil
	}

	result := &ExchangeImportPrepareResult{
		Mode:     "select_account",
		Accounts: make([]ExchangeImportAccountSummary, 0, len(accounts)),
	}
	for _, account := range accounts {
		result.Accounts = append(result.Accounts, ExchangeImportAccountSummary{ID: account.ID, Name: account.AccountName})
	}

	if len(accounts) == 1 {
		result.Mode = "single_account"
		result.AutoSelectAccountID = accounts[0].ID
	}

	return result, nil
}

func (s *PositionService) DecryptAccountCredentials(account *models.ExchangeAccount) (*ExchangeImportCredentials, error) {
	if account == nil {
		return nil, fmt.Errorf("exchange account is empty")
	}

	if strings.TrimSpace(account.DekEnc) == "" {
		creds := &ExchangeImportCredentials{
			APIKey:    strings.TrimSpace(account.ApiKey),
			SecretKey: strings.TrimSpace(account.SecretKey),
		}
		if account.AddKey != nil {
			creds.AddKey = strings.TrimSpace(*account.AddKey)
		}
		if creds.APIKey == "" || creds.SecretKey == "" {
			return nil, fmt.Errorf("exchange account credentials are incomplete")
		}
		return creds, nil
	}

	dek, err := decryptDEKWithHSM(context.Background(), account.DekEnc, account.EncAlg, account.EncKeyVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt exchange account DEK: %w", err)
	}

	apiKey, err := decryptWithDEK(dek, account.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt api key: %w", err)
	}
	secretKey, err := decryptWithDEK(dek, account.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret key: %w", err)
	}

	addKey := ""
	if account.AddKey != nil {
		addKey, err = decryptWithDEK(dek, *account.AddKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt additional key: %w", err)
		}
	}

	apiKey = strings.TrimSpace(apiKey)
	secretKey = strings.TrimSpace(secretKey)
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("exchange account credentials are incomplete")
	}

	return &ExchangeImportCredentials{
		APIKey:    apiKey,
		SecretKey: secretKey,
		AddKey:    strings.TrimSpace(addKey),
	}, nil
}

func (s *PositionService) PreviewExchangeImport(userID int, userTimezone string, positionID, exchangeAccountID int) (*ExchangeImportPreviewResult, error) {
	if positionID <= 0 {
		return nil, fmt.Errorf("position id is invalid")
	}
	if exchangeAccountID <= 0 {
		return nil, fmt.Errorf("exchange account id is invalid")
	}

	position, err := s.repo.GetPositionExchangeAndContract(userID, positionID)
	if err != nil {
		return nil, err
	}
	if position == nil {
		return nil, fmt.Errorf("position data is empty")
	}

	accountRepo := repositories.NewExchangeAccountRepository()
	account, err := accountRepo.FindByID(exchangeAccountID, userID)
	if err != nil {
		return nil, fmt.Errorf("exchange account is not found")
	}
	if account.ExID != position.ExchangeID {
		return nil, fmt.Errorf("selected exchange account does not match position exchange")
	}
	if !account.IsActive() || !account.ExchangeActive {
		return nil, fmt.Errorf("selected exchange account is not active")
	}

	credentials, err := s.DecryptAccountCredentials(account)
	if err != nil {
		return nil, err
	}

	startUTC, err := s.repo.GetLastTransactionDate(positionID, userID)
	if err != nil {
		return nil, err
	}
	if startUTC == nil {
		opened := position.OpenedAtUTC.UTC()
		startUTC = &opened
	}

	provider := s.getHistoryProvider(position.ExchangeName)
	rows, err := provider.Fetch(context.Background(), FetchHistoryRequest{
		ExchangeName: position.ExchangeName,
		MarketType:   position.MarketType,
		Contract:     position.ContractName,
		StartUTC:     startUTC,
		EndUTC:       nil,
		Credentials:  *credentials,
	})
	if err != nil {
		if errors.Is(err, errExchangeHistoryProviderNotImplemented) {
			return &ExchangeImportPreviewResult{Warning: "service is not implemented"}, nil
		}
		return nil, err
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].TransDateUTC.Before(rows[j].TransDateUTC)
	})

	token, err := defaultExchangeImportPreviewStore.save(userID, positionID, rows)
	if err != nil {
		return nil, err
	}

	previewRows := make([]ExchangeImportPreviewRow, 0, len(rows))
	for index := len(rows) - 1; index >= 0; index-- {
		item := rows[index]
		previewRows = append(previewRows, ExchangeImportPreviewRow{
			ID:               strconv.Itoa(index + 1),
			Type:             item.Type,
			Price:            item.Price,
			Volume:           item.Volume,
			FeeBaseCurrency:  item.FeeBase,
			FeeQuoteCurrency: item.FeeQuote,
			Funding:          item.Funding,
			TransactionDate:  s.formatTimeInUserTZ(item.TransDateUTC, userTimezone),
		})
	}

	return &ExchangeImportPreviewResult{
		PreviewToken: token,
		Rows:         previewRows,
	}, nil
}

func (s *PositionService) ImportExchangeTransactions(userID, positionID int, previewToken string, selectedIDs []string) (*ExchangeImportRowsResult, error) {
	if positionID <= 0 {
		return nil, fmt.Errorf("position id is invalid")
	}
	previewToken = strings.TrimSpace(previewToken)
	if previewToken == "" {
		return nil, fmt.Errorf("preview token is empty")
	}
	if len(selectedIDs) == 0 {
		return nil, fmt.Errorf("no rows selected")
	}

	entry, ok := defaultExchangeImportPreviewStore.load(previewToken)
	if !ok {
		return nil, fmt.Errorf("preview token is not found or expired")
	}
	if entry.UserID != userID || entry.PositionID != positionID {
		return nil, fmt.Errorf("preview token does not belong to this user or position")
	}

	selectedSet := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		clean := strings.TrimSpace(id)
		if clean != "" {
			selectedSet[clean] = struct{}{}
		}
	}
	if len(selectedSet) == 0 {
		return nil, fmt.Errorf("no rows selected")
	}

	result := &ExchangeImportRowsResult{}
	for index, row := range entry.Rows {
		rowID := strconv.Itoa(index + 1)
		if _, exists := selectedSet[rowID]; !exists {
			continue
		}
		result.TotalSelected++

		typeUpper := strings.ToUpper(strings.TrimSpace(row.Type))
		switch typeUpper {
		case "TRADE":
			inserted, err := s.repo.InsertTradeTransactionImport(positionID, row.Price, row.Volume, row.FeeQuote, row.FeeBase, row.TransDateUTC, row.SourceOrderID, row.SourceTradeID)
			if err != nil {
				return nil, err
			}
			if inserted {
				result.Inserted++
			} else {
				result.Skipped++
			}
		case "FUNDING":
			inserted, err := s.repo.InsertFundingTransactionImport(positionID, row.Funding, row.TransDateUTC, row.SourceOrderID, row.SourceTradeID)
			if err != nil {
				return nil, err
			}
			if inserted {
				result.Inserted++
			} else {
				result.Skipped++
			}
		default:
			result.Skipped++
		}
	}

	defaultExchangeImportPreviewStore.delete(previewToken)
	return result, nil
}

func (s *PositionService) formatTimeInUserTZ(value time.Time, userTimezone string) string {
	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		loc = time.UTC
	}
	return value.In(loc).Format(dateTimeFormat)
}

func (p *BybitHistoryProvider) Fetch(ctx context.Context, req FetchHistoryRequest) ([]exchangeHistoryCandidate, error) {
	if strings.TrimSpace(req.Credentials.APIKey) == "" || strings.TrimSpace(req.Credentials.SecretKey) == "" {
		return nil, fmt.Errorf("exchange api credentials are not configured")
	}

	category := "spot"
	if strings.EqualFold(strings.TrimSpace(req.MarketType), "FUTURES") {
		category = "linear"
	}

	symbol := normalizeBybitSymbol(req.Contract)
	if symbol == "" {
		return nil, fmt.Errorf("contract is empty")
	}

	cursor := ""
	all := make([]exchangeHistoryCandidate, 0)
	for page := 0; page < bybitHistoryMaxPages; page++ {
		params := url.Values{}
		params.Set("category", category)
		params.Set("symbol", symbol)
		params.Set("limit", strconv.Itoa(bybitHistoryPageSize))
		if req.StartUTC != nil {
			params.Set("startTime", strconv.FormatInt(req.StartUTC.UnixMilli(), 10))
		}
		if req.EndUTC != nil {
			params.Set("endTime", strconv.FormatInt(req.EndUTC.UnixMilli(), 10))
		}
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		payload, err := p.bybitGet(ctx, "/v5/execution/list", params, req.Credentials)
		if err != nil {
			return nil, err
		}

		var response bybitExecutionResponse
		if err := json.Unmarshal(payload, &response); err != nil {
			return nil, fmt.Errorf("decode bybit response: %w", err)
		}
		if response.RetCode != 0 {
			return nil, fmt.Errorf("bybit api error: %s", strings.TrimSpace(response.RetMsg))
		}

		for _, item := range response.Result.List {
			if normalizeBybitSymbol(item.Symbol) != symbol {
				continue
			}

			execType := strings.ToUpper(strings.TrimSpace(item.ExecType))
			if execType != "TRADE" && execType != "FUNDING" {
				continue
			}

			timeMS, err := strconv.ParseInt(strings.TrimSpace(item.ExecTime), 10, 64)
			if err != nil {
				continue
			}
			transDateUTC := time.UnixMilli(timeMS).UTC()
			// Strict lower bound: preview should include only records newer than the latest imported one.
			if req.StartUTC != nil && !transDateUTC.After(*req.StartUTC) {
				continue
			}
			if req.EndUTC != nil && transDateUTC.After(*req.EndUTC) {
				continue
			}

			var sourceOrderID *string
			orderID := strings.TrimSpace(item.OrderID)
			if orderID != "" {
				sourceOrderID = &orderID
			}
			var sourceTradeID *string
			execID := strings.TrimSpace(item.ExecID)
			if execID != "" {
				sourceTradeID = &execID
			}

			if execType == "FUNDING" {
				// Bybit returns funding in execution fee semantics (expense-positive),
				// while POS_TRANSACTIONS stores PnL semantics (income-positive).
				fundingRaw := negateDecimalString(strings.TrimSpace(item.ExecFee))
				if strings.TrimSpace(fundingRaw) == "" {
					fundingRaw = strings.TrimSpace(item.ExecValue)
				}
				funding, normErr := normalizeCSVDecimal(fundingRaw)
				if normErr != nil {
					continue
				}
				all = append(all, exchangeHistoryCandidate{
					Type:          "FUNDING",
					Price:         "0",
					Volume:        "0",
					FeeBase:       "0",
					FeeQuote:      "0",
					Funding:       funding,
					TransDateUTC:  transDateUTC,
					SourceOrderID: sourceOrderID,
					SourceTradeID: sourceTradeID,
				})
				continue
			}

			price, err := normalizeCSVDecimal(item.ExecPrice)
			if err != nil {
				continue
			}
			qty, err := normalizeCSVDecimal(item.ExecQty)
			if err != nil {
				continue
			}
			fee, err := normalizeCSVDecimal(item.ExecFee)
			if err != nil {
				continue
			}

			volume := absDecimalString(qty)
			if strings.EqualFold(strings.TrimSpace(item.Side), "Sell") && !isZeroDecimal(volume) {
				volume = "-" + volume
			}

			all = append(all, exchangeHistoryCandidate{
				Type:          "TRADE",
				Price:         absDecimalString(price),
				Volume:        volume,
				FeeBase:       "0",
				FeeQuote:      absDecimalString(fee),
				Funding:       "0",
				TransDateUTC:  transDateUTC,
				SourceOrderID: sourceOrderID,
				SourceTradeID: sourceTradeID,
			})
		}

		cursor = strings.TrimSpace(response.Result.NextPageCursor)
		if cursor == "" {
			break
		}
	}

	return all, nil
}

func (p *BybitHistoryProvider) bybitGet(ctx context.Context, path string, queryParams url.Values, credentials ExchangeImportCredentials) ([]byte, error) {
	if p.httpClient == nil {
		p.httpClient = &http.Client{Timeout: exchangeImportHTTPCallTimeout}
	}

	queryString := queryParams.Encode()
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signPayload := timestamp + credentials.APIKey + bybitRecvWindow + queryString
	signature := signBybitPayload(signPayload, credentials.SecretKey)

	endpoint := bybitAPIBaseURL + path
	if queryString != "" {
		endpoint += "?" + queryString
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build bybit request: %w", err)
	}
	httpReq.Header.Set("X-BAPI-API-KEY", credentials.APIKey)
	httpReq.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-BAPI-RECV-WINDOW", bybitRecvWindow)
	httpReq.Header.Set("X-BAPI-SIGN", signature)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request bybit api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bybit response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bybit http status %d", resp.StatusCode)
	}

	return body, nil
}

func signBybitPayload(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func normalizeBybitSymbol(value string) string {
	clean := strings.ToUpper(strings.TrimSpace(value))
	clean = strings.ReplaceAll(clean, "/", "")
	clean = strings.ReplaceAll(clean, "-", "")
	return clean
}

func negateDecimalString(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "-") {
		return strings.TrimPrefix(value, "-")
	}
	value = strings.TrimPrefix(value, "+")
	return "-" + value
}
