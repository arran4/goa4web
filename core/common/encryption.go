package common

import (
	"crypto/sha256"
	"fmt"

	"github.com/gorilla/securecookie"
)

// EncryptData encrypts and signs the given string using the session secret.
func (cd *CoreData) EncryptData(data string) (string, error) {
	if cd.Config == nil || cd.Config.SessionSecret == "" {
		return "", fmt.Errorf("session secret missing")
	}
	sc := getSecureCookie(cd.Config.SessionSecret)
	encoded, err := sc.Encode("d", data)
	if err != nil {
		return "", err
	}
	return encoded, nil
}

// DecryptData verifies and decrypts the given string.
func (cd *CoreData) DecryptData(encoded string) (string, error) {
	if cd.Config == nil || cd.Config.SessionSecret == "" {
		return "", fmt.Errorf("session secret missing")
	}
	sc := getSecureCookie(cd.Config.SessionSecret)
	var data string
	if err := sc.Decode("d", encoded, &data); err != nil {
		return "", err
	}
	return data, nil
}

func getSecureCookie(secret string) *securecookie.SecureCookie {
	// Derive keys from the single session secret to ensure they are the correct length.
	// Hash key: 32 bytes (SHA256)
	// Block key: 32 bytes (SHA256 of the secret reversed? or just hashed again)

	// Simple derivation:
	// Hash key = SHA256(secret)
	// Block key = SHA256(secret + "block")

	h := sha256.New()
	h.Write([]byte(secret))
	hashKey := h.Sum(nil)

	h.Reset()
	h.Write([]byte(secret))
	h.Write([]byte("block"))
	blockKey := h.Sum(nil)

	return securecookie.New(hashKey, blockKey)
}
