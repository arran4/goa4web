package adminapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Signer signs and verifies admin API requests.
type Signer struct{ key string }

// NewSigner returns a Signer using the provided key.
func NewSigner(key string) *Signer { return &Signer{key: key} }

func (s *Signer) signString(method, path string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(s.key))
	io.WriteString(mac, fmt.Sprintf("%s:%s:%d", method, path, exp))
	return hex.EncodeToString(mac.Sum(nil))
}

// Sign computes a timestamped signature for the request method and path.
func (s *Signer) Sign(method, path string) (int64, string) {
	exp := time.Now().Add(5 * time.Minute).Unix()
	return exp, s.signString(method, path, exp)
}

// Verify checks that tsStr and sig match the provided method and path.
func (s *Signer) Verify(method, path, tsStr, sig string) bool {
	exp, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	want := s.signString(method, path, exp)
	return hmac.Equal([]byte(want), []byte(sig))
}
