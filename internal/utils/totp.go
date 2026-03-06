package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	totpDefaultDigits = 6
	totpDefaultPeriod = 30
)

// GenerateTOTPSecret returns a random base32-encoded secret.
func GenerateTOTPSecret(length int) (string, error) {
	if length <= 0 {
		length = 20
	}

	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)
	return strings.ToUpper(encoder.EncodeToString(raw)), nil
}

// BuildOTPAuthURI creates otpauth URI for authenticator apps.
func BuildOTPAuthURI(secret, issuer, account string) string {
	if issuer == "" {
		issuer = "CT-System"
	}
	label := url.QueryEscape(issuer + ":" + account)
	return fmt.Sprintf(
		"otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		label,
		url.QueryEscape(secret),
		url.QueryEscape(issuer),
		totpDefaultDigits,
		totpDefaultPeriod,
	)
}

// VerifyTOTPCode verifies TOTP code with a configurable time drift window.
func VerifyTOTPCode(secret, code string, now time.Time, window int) bool {
	cleanCode := strings.TrimSpace(code)
	if len(cleanCode) != totpDefaultDigits {
		return false
	}
	if _, err := strconv.Atoi(cleanCode); err != nil {
		return false
	}

	counter := now.Unix() / totpDefaultPeriod
	for i := -window; i <= window; i++ {
		if generateTOTPAtCounter(secret, counter+int64(i)) == cleanCode {
			return true
		}
	}
	return false
}

func generateTOTPAtCounter(secret string, counter int64) string {
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return ""
	}

	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(counter))

	h := hmac.New(sha1.New, secretBytes)
	h.Write(counterBytes)
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binaryCode := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)

	otp := binaryCode % 1000000
	return fmt.Sprintf("%06d", otp)
}
