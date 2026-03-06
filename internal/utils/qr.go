package utils

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCodeDataURI creates a PNG QR code as data URI for embedding in HTML.
func GenerateQRCodeDataURI(content string, size int) (string, error) {
	if size <= 0 {
		size = 256
	}

	pngBytes, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return "", fmt.Errorf("failed to generate qr code: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}
