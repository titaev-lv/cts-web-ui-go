// Package services содержит бизнес-логику для работы с биржами и их аккаунтами.
package services

import (
	"context"
	"ctweb/internal/models"
	"ctweb/internal/repositories"
	"ctweb/internal/utils"
	"errors"
	"fmt"
	"strings"
)

func isUnrecoverableDEKError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "decryption failed") || strings.Contains(msg, "hsm decrypt error")
}

// ExchangeService инкапсулирует валидацию и бизнес-логику для бирж и аккаунтов бирж.
type ExchangeService struct {
	exchangeRepo *repositories.ExchangeRepository
	accountRepo  *repositories.ExchangeAccountRepository
}

// NewExchangeService создаёт сервис с необходимыми репозиториями.
func NewExchangeService() *ExchangeService {
	return &ExchangeService{
		exchangeRepo: repositories.NewExchangeRepository(),
		accountRepo:  repositories.NewExchangeAccountRepository(),
	}
}

// ExchangeRepo возвращает репозиторий бирж (используется в контроллере для DataTables).
func (s *ExchangeService) ExchangeRepo() *repositories.ExchangeRepository {
	return s.exchangeRepo
}

// AccountRepo возвращает репозиторий аккаунтов бирж.
func (s *ExchangeService) AccountRepo() *repositories.ExchangeAccountRepository {
	return s.accountRepo
}

// =============================
// Helpers: status
// =============================

// normalizeStatus проверяет статус (enable/disable) и возвращает bool Active.
func normalizeStatus(status string) (bool, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if err := utils.ValidateStatus(status); err != nil {
		return false, err
	}
	return utils.StatusToBool(status), nil
}

// normalizeAccountStatus проверяет статус для аккаунта (Active/Blocked) и возвращает bool Active.
func normalizeAccountStatus(status string) (bool, error) {
	status = strings.TrimSpace(status)
	if status == "Active" {
		return true, nil
	}
	if status == "Blocked" {
		return false, nil
	}
	return false, errors.New("invalid status")
}

// normalizeExchangeStatus проверяет статус для биржи (Active/Blocked) и возвращает bool Active.
func normalizeExchangeStatus(status string) (bool, error) {
	status = strings.TrimSpace(status)
	if status == "Active" {
		return true, nil
	}
	if status == "Blocked" {
		return false, nil
	}
	return false, errors.New("invalid status")
}

// =============================
// Exchange (биржи)
// =============================

// ValidateExchange проверяет обязательные поля и статус.
func (s *ExchangeService) ValidateExchange(name, url, baseURL, classToFactory, status string) (bool, error) {
	if strings.TrimSpace(name) == "" {
		return false, errors.New("name is required")
	}
	if strings.TrimSpace(url) == "" {
		return false, errors.New("url is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		return false, errors.New("base_url is required")
	}
	if strings.TrimSpace(classToFactory) == "" {
		return false, errors.New("class is required")
	}

	active, err := normalizeExchangeStatus(status)
	if err != nil {
		return false, err
	}
	return active, nil
}

// EnsureExchangeNameUnique проверяет уникальность имени (при создании/обновлении).
func (s *ExchangeService) EnsureExchangeNameUnique(name string, excludeID *int) error {
	if excludeID == nil {
		exists, err := s.exchangeRepo.ExistsByName(name)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("exchange name already exists")
		}
		return nil
	}

	exists, err := s.exchangeRepo.ExistsByNameExcludingID(name, *excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("exchange name already exists")
	}
	return nil
}

// CreateExchange валидирует данные, проверяет уникальность и создаёт запись.
func (s *ExchangeService) CreateExchange(name, url, baseURL, classToFactory, status string, description *string, websocketURL *string, userID int) (int, error) {
	active, err := s.ValidateExchange(name, url, baseURL, classToFactory, status)
	if err != nil {
		return 0, err
	}
	if err := s.EnsureExchangeNameUnique(name, nil); err != nil {
		return 0, err
	}

	ex := &models.Exchange{
		Name:           strings.TrimSpace(name),
		URL:            strings.TrimSpace(url),
		BaseURL:        strings.TrimSpace(baseURL),
		ClassToFactory: strings.TrimSpace(classToFactory),
		Active:         active,
		Deleted:        false,
		Description:    description,
		WebsocketURL:   websocketURL,
		UserCreated:    &userID,
	}
	return s.exchangeRepo.Create(ex, userID)
}

// UpdateExchange валидирует данные, проверяет уникальность и обновляет запись.
func (s *ExchangeService) UpdateExchange(id int, name, url, baseURL, classToFactory, status string, description *string, websocketURL *string, userID int) error {
	active, err := s.ValidateExchange(name, url, baseURL, classToFactory, status)
	if err != nil {
		return err
	}
	if err := s.EnsureExchangeNameUnique(name, &id); err != nil {
		return err
	}

	ex := &models.Exchange{
		ID:             id,
		Name:           strings.TrimSpace(name),
		URL:            strings.TrimSpace(url),
		BaseURL:        strings.TrimSpace(baseURL),
		ClassToFactory: strings.TrimSpace(classToFactory),
		Active:         active,
		Deleted:        false,
		Description:    description,
		WebsocketURL:   websocketURL,
		UserModify:     &userID,
	}
	return s.exchangeRepo.Update(ex, userID)
}

// =============================
// Exchange Accounts (аккаунты бирж)
// =============================

// ValidateExchangeAccount проверяет обязательные поля и статус.
func (s *ExchangeService) ValidateExchangeAccount(accountName, status string, priority int, apiKey string) (bool, error) {
	if strings.TrimSpace(accountName) == "" {
		return false, errors.New("account name is required")
	}
	if strings.TrimSpace(apiKey) == "" {
		return false, errors.New("api key is required")
	}
	// priority может быть 0 и выше
	if priority < 0 {
		return false, errors.New("priority must be >= 0")
	}

	active, err := normalizeAccountStatus(status)
	if err != nil {
		return false, err
	}
	return active, nil
}

// EnsureExchangeAccountNameUnique проверяет уникальность имени аккаунта в рамках пользователя и биржи.
func (s *ExchangeService) EnsureExchangeAccountNameUnique(userID, exchangeID int, name string, excludeID *int) error {
	if excludeID == nil {
		exists, err := s.accountRepo.ExistsByName(userID, exchangeID, name)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("exchange account name already exists")
		}
		return nil
	}

	exists, err := s.accountRepo.ExistsByNameExcludingID(userID, exchangeID, *excludeID, name)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("exchange account name already exists")
	}
	return nil
}

// CreateExchangeAccount валидирует и создаёт аккаунт.
func (s *ExchangeService) CreateExchangeAccount(userID, exchangeID int, accountName, status string, priority int, apiKey, secretKey, addKey, note string) (int, error) {
	active, err := s.ValidateExchangeAccount(accountName, status, priority, apiKey)
	if err != nil {
		return 0, err
	}
	if err := s.EnsureExchangeAccountNameUnique(userID, exchangeID, accountName, nil); err != nil {
		return 0, err
	}

	dek, err := newRandomDEK()
	if err != nil {
		return 0, err
	}
	dekEnc, encKeyVersion, encAlg, err := encryptDEKWithHSM(context.Background(), dek)
	if err != nil {
		return 0, err
	}

	encryptedAPIKey, err := encryptWithDEK(dek, strings.TrimSpace(apiKey))
	if err != nil {
		return 0, err
	}
	encryptedSecretKey, err := encryptWithDEK(dek, strings.TrimSpace(secretKey))
	if err != nil {
		return 0, err
	}
	encryptedAddKey, err := encryptWithDEK(dek, strings.TrimSpace(addKey))
	if err != nil {
		return 0, err
	}

	acc := &models.ExchangeAccount{
		ExID:          exchangeID,
		UID:           userID,
		AccountName:   strings.TrimSpace(accountName),
		Priority:      priority,
		Active:        active,
		ApiKey:        encryptedAPIKey,
		SecretKey:     encryptedSecretKey,
		DekEnc:        dekEnc,
		EncKeyVersion: encKeyVersion,
		EncAlg:        encAlg,
	}

	if strings.TrimSpace(encryptedAddKey) != "" {
		acc.AddKey = &encryptedAddKey
	}

	trimmedNote := strings.TrimSpace(note)
	if trimmedNote != "" {
		acc.Note = &trimmedNote
	}

	return s.accountRepo.Create(acc)
}

// UpdateExchangeAccount валидирует и обновляет аккаунт.
func (s *ExchangeService) UpdateExchangeAccount(id, userID, exchangeID int, accountName, status string, priority int, apiKey, secretKey, addKey, note string) error {
	if strings.TrimSpace(accountName) == "" {
		return errors.New("account name is required")
	}
	if priority < 0 {
		return errors.New("priority must be >= 0")
	}

	active, err := normalizeAccountStatus(status)
	if err != nil {
		return err
	}
	if err := s.EnsureExchangeAccountNameUnique(userID, exchangeID, accountName, &id); err != nil {
		return err
	}

	existing, err := s.accountRepo.FindByID(id, userID)
	if err != nil {
		return err
	}

	apiProvided := strings.TrimSpace(apiKey) != ""
	secretProvided := strings.TrimSpace(secretKey) != ""
	addProvided := strings.TrimSpace(addKey) != ""

	finalAPIEnc := ""
	finalSecretEnc := ""
	var finalAddEnc *string

	encMetaChanged := false
	dekEnc := existing.DekEnc
	encKeyVersion := existing.EncKeyVersion
	encAlg := existing.EncAlg

	if apiProvided || secretProvided || addProvided {
		if strings.TrimSpace(existing.DekEnc) != "" {
			dek, dekErr := decryptDEKWithHSM(context.Background(), existing.DekEnc, existing.EncAlg, existing.EncKeyVersion)
			if dekErr != nil {
				if !isUnrecoverableDEKError(dekErr) {
					return fmt.Errorf("failed to decrypt credentials DEK: %w", dekErr)
				}

				// Old KEK is unavailable. Allow forced re-key with newly provided credentials.
				if !apiProvided {
					return errors.New("stored credentials are encrypted with a retired KEK and cannot be decrypted; provide a new API key to reset credentials")
				}

				newDEK, genErr := newRandomDEK()
				if genErr != nil {
					return genErr
				}
				newDekEnc, newEncKeyVersion, newEncAlg, wrapErr := encryptDEKWithHSM(context.Background(), newDEK)
				if wrapErr != nil {
					return wrapErr
				}

				encAPI, encErr := encryptWithDEK(newDEK, strings.TrimSpace(apiKey))
				if encErr != nil {
					return encErr
				}
				encSecret, encErr := encryptWithDEK(newDEK, strings.TrimSpace(secretKey))
				if encErr != nil {
					return encErr
				}
				encAdd, encErr := encryptWithDEK(newDEK, strings.TrimSpace(addKey))
				if encErr != nil {
					return encErr
				}

				finalAPIEnc = encAPI
				finalSecretEnc = encSecret
				if strings.TrimSpace(encAdd) != "" {
					finalAddEnc = &encAdd
				}

				dekEnc = newDekEnc
				encKeyVersion = newEncKeyVersion
				encAlg = newEncAlg
				encMetaChanged = true
			} else {
				if apiProvided {
					enc, encErr := encryptWithDEK(dek, strings.TrimSpace(apiKey))
					if encErr != nil {
						return encErr
					}
					finalAPIEnc = enc
				}
				if secretProvided {
					enc, encErr := encryptWithDEK(dek, strings.TrimSpace(secretKey))
					if encErr != nil {
						return encErr
					}
					finalSecretEnc = enc
				}
				if addProvided {
					enc, encErr := encryptWithDEK(dek, strings.TrimSpace(addKey))
					if encErr != nil {
						return encErr
					}
					if strings.TrimSpace(enc) != "" {
						finalAddEnc = &enc
					}
				}
			}
		} else {
			// Legacy row without DEK: treat existing values as plaintext and migrate to envelope encryption.
			plainAPI := strings.TrimSpace(apiKey)
			if plainAPI == "" {
				plainAPI = strings.TrimSpace(existing.ApiKey)
			}
			if plainAPI == "" {
				return errors.New("api key is required for credential encryption migration")
			}

			plainSecret := strings.TrimSpace(secretKey)
			if plainSecret == "" {
				plainSecret = strings.TrimSpace(existing.SecretKey)
			}

			plainAdd := strings.TrimSpace(addKey)
			if plainAdd == "" && existing.AddKey != nil {
				plainAdd = strings.TrimSpace(*existing.AddKey)
			}

			dek, dekErr := newRandomDEK()
			if dekErr != nil {
				return dekErr
			}
			newDekEnc, newEncKeyVersion, newEncAlg, encErr := encryptDEKWithHSM(context.Background(), dek)
			if encErr != nil {
				return encErr
			}

			encAPI, encErr := encryptWithDEK(dek, plainAPI)
			if encErr != nil {
				return encErr
			}
			encSecret, encErr := encryptWithDEK(dek, plainSecret)
			if encErr != nil {
				return encErr
			}
			encAdd, encErr := encryptWithDEK(dek, plainAdd)
			if encErr != nil {
				return encErr
			}

			finalAPIEnc = encAPI
			finalSecretEnc = encSecret
			if strings.TrimSpace(encAdd) != "" {
				finalAddEnc = &encAdd
			}

			dekEnc = newDekEnc
			encKeyVersion = newEncKeyVersion
			encAlg = newEncAlg
			encMetaChanged = true
		}
	}

	acc := &models.ExchangeAccount{
		ID:          id,
		ExID:        exchangeID,
		UID:         userID,
		AccountName: strings.TrimSpace(accountName),
		Priority:    priority,
		Active:      active,
		ApiKey:      finalAPIEnc,
		SecretKey:   finalSecretEnc,
	}
	if encMetaChanged {
		acc.DekEnc = dekEnc
		acc.EncKeyVersion = encKeyVersion
		acc.EncAlg = encAlg
	}

	if finalAddEnc != nil {
		acc.AddKey = finalAddEnc
	}

	trimmedNote := strings.TrimSpace(note)
	if trimmedNote != "" {
		acc.Note = &trimmedNote
	}

	return s.accountRepo.Update(acc)
}

// SoftDeleteExchangeAccount помечает аккаунт удалённым.
func (s *ExchangeService) SoftDeleteExchangeAccount(id, userID int) error {
	return s.accountRepo.SoftDelete(id, userID)
}

// =============================
// Дополнительные проверки
// =============================

// ValidateExchangeExists проверяет наличие биржи (для случаев, когда нужно убедиться, что EXID валиден).
func (s *ExchangeService) ValidateExchangeExists(exchangeID int) error {
	_, err := s.exchangeRepo.FindByID(exchangeID)
	if err != nil {
		return fmt.Errorf("exchange not found: %w", err)
	}
	return nil
}
