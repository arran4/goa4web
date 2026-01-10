package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Signer signs and verifies arbitrary data using HMAC-SHA256.
type Signer struct {
	Key           string
	DefaultExpiry time.Duration
}

// Sign generates a signature for data. When exp is zero, the
// signature never expires. If no timestamp is provided, a default of 24
// hours from now is used.
func (s *Signer) Sign(data string, exp ...time.Time) (int64, string) {
	var ts int64
	if len(exp) > 0 {
		ts = exp[0].Unix()
	} else {
		expiry := 24 * time.Hour
		if s.DefaultExpiry > 0 {
			expiry = s.DefaultExpiry
		}
		ts = time.Now().Add(expiry).Unix()
	}
	mac := hmac.New(sha256.New, []byte(s.Key))
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, ts))
	return ts, hex.EncodeToString(mac.Sum(nil))
}

// Verify checks data against ts and sig. If ts is zero, the signature
// is considered valid indefinitely.
func (s *Signer) Verify(data, tsStr, sig string) bool {
	exp, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return false
	}
	if exp != 0 && time.Now().Unix() > exp {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.Key))
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, exp))
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(sig))
}
