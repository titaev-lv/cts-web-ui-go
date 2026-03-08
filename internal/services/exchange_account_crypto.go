package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"ctweb/internal/config"
	"ctweb/internal/hsm"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	exchangeCredEncPrefix = "v1:"
	exchangeCredEncAlg    = "AES-256-GCM"
)

var (
	exchangeHSMClientOnce sync.Once
	exchangeHSMClientInst *hsm.Client
	exchangeHSMClientErr  error
)

func getExchangeHSMClient() (*hsm.Client, string, error) {
	cfg := config.Get()
	if !cfg.HSM.Enabled {
		return nil, "", fmt.Errorf("hsm is disabled")
	}

	exchangeHSMClientOnce.Do(func() {
		exchangeHSMClientInst, exchangeHSMClientErr = hsm.NewClient(hsm.ClientConfig{
			BaseURL:        cfg.HSM.URL,
			CertPath:       cfg.HSM.Trading.TLS.CertPath,
			KeyPath:        cfg.HSM.Trading.TLS.KeyPath,
			CAPath:         cfg.HSM.Trading.TLS.CAPath,
			RequestTimeout: cfg.HSM.Timeout,
			RetryConfig: hsm.RetryConfig{
				MaxAttempts: cfg.HSM.Retry.MaxAttempts,
				InitialWait: cfg.HSM.Retry.InitialDelay,
				MaxWait:     cfg.HSM.Retry.MaxDelay,
				Multiplier:  cfg.HSM.Retry.Multiplier,
			},
		})
	})

	if exchangeHSMClientErr != nil {
		return nil, "", exchangeHSMClientErr
	}

	ctxName := strings.TrimSpace(cfg.HSM.Trading.Context)
	if ctxName == "" {
		ctxName = "exchange-key"
	}
	return exchangeHSMClientInst, ctxName, nil
}

func newRandomDEK() ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	return dek, nil
}

func encryptWithDEK(dek []byte, plaintext string) (string, error) {
	if strings.TrimSpace(plaintext) == "" {
		return "", nil
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", fmt.Errorf("failed to init aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to init gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	combined := append(nonce, sealed...)
	return exchangeCredEncPrefix + base64.StdEncoding.EncodeToString(combined), nil
}

func decryptWithDEK(dek []byte, ciphertext string) (string, error) {
	value := strings.TrimSpace(ciphertext)
	if value == "" {
		return "", nil
	}
	value = strings.TrimPrefix(value, exchangeCredEncPrefix)

	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(dek)
	if err != nil {
		return "", fmt.Errorf("failed to init aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to init gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	nonce := raw[:nonceSize]
	payload := raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return string(plain), nil
}

func parseHSMKeyVersionFromKeyID(keyID string) int {
	re := regexp.MustCompile(`v(\d+)$`)
	m := re.FindStringSubmatch(strings.TrimSpace(keyID))
	if len(m) != 2 {
		return 0
	}
	v, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return v
}

func keyIDFromEncMeta(encAlg string, encKeyVersion int) string {
	parts := strings.Split(strings.TrimSpace(encAlg), "|")
	if len(parts) >= 2 {
		keyID := strings.TrimSpace(parts[len(parts)-1])
		if keyID != "" {
			return keyID
		}
	}

	if encKeyVersion > 0 {
		return "kek-exchange-key-v" + strconv.Itoa(encKeyVersion)
	}
	return "kek-exchange-key-v1"
}

func encryptDEKWithHSM(ctx context.Context, dek []byte) (dekEnc string, encKeyVersion int, encAlg string, err error) {
	client, hsmContext, err := getExchangeHSMClient()
	if err != nil {
		return "", 0, "", err
	}

	keyID, cipherText, err := client.Encrypt(ctx, hsmContext, dek)
	if err != nil {
		return "", 0, "", err
	}

	version := parseHSMKeyVersionFromKeyID(keyID)
	if version <= 0 {
		version = 1
	}

	return cipherText, version, exchangeCredEncAlg + "|" + keyID, nil
}

func decryptDEKWithHSM(ctx context.Context, dekEnc, encAlg string, encKeyVersion int) ([]byte, error) {
	client, hsmContext, err := getExchangeHSMClient()
	if err != nil {
		return nil, err
	}

	keyID := keyIDFromEncMeta(encAlg, encKeyVersion)
	plain, err := client.Decrypt(ctx, hsmContext, keyID, dekEnc)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
