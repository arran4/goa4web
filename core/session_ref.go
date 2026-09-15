package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func NewSessionRef() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashSessionRef(ref string) string {
	h := sha256.Sum256([]byte(ref))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
