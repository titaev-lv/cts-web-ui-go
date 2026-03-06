package hsm

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

// RetryConfig configures retry behavior for HSM requests.
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// ClientConfig configures HSM client.
type ClientConfig struct {
	BaseURL        string
	CertPath       string
	KeyPath        string
	CAPath         string
	RequestTimeout time.Duration
	RetryConfig    RetryConfig
}

// Client talks to HSM service over mTLS.
type Client struct {
	baseURL    string
	httpClient *http.Client
	retryCfg   RetryConfig
}

// NewClient creates a new HSM client.
func NewClient(cfg ClientConfig) (*Client, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	caCert, err := os.ReadFile(cfg.CAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
	}

	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.RetryConfig.MaxAttempts <= 0 {
		cfg.RetryConfig.MaxAttempts = 1
	}
	if cfg.RetryConfig.InitialWait <= 0 {
		cfg.RetryConfig.InitialWait = 250 * time.Millisecond
	}
	if cfg.RetryConfig.MaxWait <= 0 {
		cfg.RetryConfig.MaxWait = 2 * time.Second
	}
	if cfg.RetryConfig.Multiplier < 1 {
		cfg.RetryConfig.Multiplier = 2
	}

	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.RequestTimeout,
		},
		retryCfg: cfg.RetryConfig,
	}, nil
}

// Close closes idle HTTP connections.
func (c *Client) Close() {
	if c != nil && c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, reqBody interface{}, out interface{}) error {
	var lastErr error

	for attempt := 1; attempt <= c.retryCfg.MaxAttempts; attempt++ {
		payload := []byte(nil)
		if reqBody != nil {
			b, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}
			payload = b
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if attempt < c.retryCfg.MaxAttempts {
				c.waitBeforeRetry(ctx, attempt)
				continue
			}
			break
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("failed to read response: %w", readErr)
			if attempt < c.retryCfg.MaxAttempts {
				c.waitBeforeRetry(ctx, attempt)
				continue
			}
			break
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("hsm http %d: %s", resp.StatusCode, string(body))
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return lastErr
			}
			if attempt < c.retryCfg.MaxAttempts {
				c.waitBeforeRetry(ctx, attempt)
				continue
			}
			break
		}

		if out != nil {
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("request failed after retries: %w", lastErr)
}

func (c *Client) waitBeforeRetry(ctx context.Context, attempt int) {
	wait := time.Duration(float64(c.retryCfg.InitialWait) * math.Pow(c.retryCfg.Multiplier, float64(attempt-1)))
	if wait > c.retryCfg.MaxWait {
		wait = c.retryCfg.MaxWait
	}

	select {
	case <-time.After(wait):
	case <-ctx.Done():
	}
}

// Encrypt encrypts plaintext bytes with HSM and returns key ID + ciphertext.
func (c *Client) Encrypt(ctx context.Context, cryptoContext string, plaintext []byte) (string, string, error) {
	req := EncryptRequest{Context: cryptoContext, Plaintext: encodeBase64(plaintext)}
	var resp EncryptResponse
	if err := c.doRequest(ctx, http.MethodPost, "/encrypt", req, &resp); err != nil {
		return "", "", err
	}
	if resp.Error != "" {
		return "", "", fmt.Errorf("hsm encrypt error: %s", resp.Error)
	}
	return resp.KeyID, resp.Ciphertext, nil
}

// Decrypt decrypts ciphertext with HSM and returns plaintext bytes.
func (c *Client) Decrypt(ctx context.Context, cryptoContext, keyID, ciphertext string) ([]byte, error) {
	req := DecryptRequest{Context: cryptoContext, KeyID: keyID, Ciphertext: ciphertext}
	var resp DecryptResponse
	if err := c.doRequest(ctx, http.MethodPost, "/decrypt", req, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("hsm decrypt error: %s", resp.Error)
	}
	plain, err := decodeBase64(resp.Plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode plaintext: %w", err)
	}
	return plain, nil
}
