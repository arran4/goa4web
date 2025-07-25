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

var signKey string

// SetSigningKey stores the key used for signing admin API requests.
func SetSigningKey(k string) { signKey = k }

func signString(method, path string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(signKey))
	io.WriteString(mac, fmt.Sprintf("%s:%s:%d", method, path, exp))
	return hex.EncodeToString(mac.Sum(nil))
}

// Sign computes a timestamped signature for the request method and path.
func Sign(method, path string) (int64, string) {
	exp := time.Now().Add(5 * time.Minute).Unix()
	return exp, signString(method, path, exp)
}

// Verify checks that tsStr and sig match the provided method and path.
func Verify(method, path, tsStr, sig string) bool {
	exp, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	want := signString(method, path, exp)
	return hmac.Equal([]byte(want), []byte(sig))
}
