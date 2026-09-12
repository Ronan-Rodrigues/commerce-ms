package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type SHA256HMACValidator struct{}

func NewSHA256HMACValidator() *SHA256HMACValidator {
	return &SHA256HMACValidator{}
}

// ValidateSignature valida a assinatura HMAC em tempo constante
func (v *SHA256HMACValidator) ValidateSignature(payload []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// hmac.Equal faz comparação em tempo constante
	return hmac.Equal(sigBytes, expectedMAC)
}
