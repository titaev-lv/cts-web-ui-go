package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const recoveryAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateRecoveryCodes creates human-readable recovery codes like XXXX-XXXX.
func GenerateRecoveryCodes(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		chunk, err := randomFromAlphabet(8)
		if err != nil {
			return nil, err
		}
		codes = append(codes, chunk[:4]+"-"+chunk[4:])
	}

	return codes, nil
}

// HashRecoveryCodes hashes recovery codes for one-way verification storage.
func HashRecoveryCodes(codes []string) []string {
	hashes := make([]string, 0, len(codes))
	for _, code := range codes {
		sum := sha256.Sum256([]byte(code))
		hashes = append(hashes, hex.EncodeToString(sum[:]))
	}
	return hashes
}

func randomFromAlphabet(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid recovery code length")
	}

	buf := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	for i := range buf {
		buf[i] = recoveryAlphabet[int(randomBytes[i])%len(recoveryAlphabet)]
	}

	return string(buf), nil
}
