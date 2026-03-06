package hsm

import "encoding/base64"

// EncryptRequest represents /encrypt request to HSM service.
type EncryptRequest struct {
	Context   string `json:"context"`
	Plaintext string `json:"plaintext"`
}

// EncryptResponse represents /encrypt response from HSM service.
type EncryptResponse struct {
	KeyID      string `json:"key_id"`
	Ciphertext string `json:"ciphertext"`
	Error      string `json:"error,omitempty"`
}

// DecryptRequest represents /decrypt request to HSM service.
type DecryptRequest struct {
	Context    string `json:"context"`
	KeyID      string `json:"key_id"`
	Ciphertext string `json:"ciphertext"`
}

// DecryptResponse represents /decrypt response from HSM service.
type DecryptResponse struct {
	Plaintext string `json:"plaintext"`
	Error     string `json:"error,omitempty"`
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
