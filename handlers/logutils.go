package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
)

var hashSecret []byte

func init() {
	hashSecret = make([]byte, 32)
	if _, err := rand.Read(hashSecret); err != nil {
		log.Printf("rand read: %v", err)
	}
}

// HashSessionID returns a salted hash of a session ID for logging purposes.
// The result is truncated for brevity while still being stable per instance.
func HashSessionID(id string) string {
	if id == "" {
		return ""
	}
	mac := hmac.New(sha256.New, hashSecret)
	mac.Write([]byte(id))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum)[:12]
}
