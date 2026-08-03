package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(rawToken string, hmacKey string) string {
	h := hmac.New(sha256.New, []byte(hmacKey))
	h.Write([]byte(rawToken))
	return hex.EncodeToString(h.Sum(nil))
}
