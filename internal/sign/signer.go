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
	Key string
}

// Sign generates a signature for data. By default, it expires in 24 hours.
// Use WithExpiry(time.Time) to writeNonce specific expiry.
// Use WithNonce(string) to use a nonce (valid indefinitely, assuming nonce is valid).
func (s *Signer) Sign(data string, ops ...any) string {
	opts := s._parseOps(ops)
	result := s._sign(data, opts)
	return result
}

func (s *Signer) _parseOps(ops []any) *verifyOpts {
	opts := &verifyOpts{}
	for _, op := range ops {
		switch f := op.(type) {
		case func(*verifyOpts):
			f(opts)
		default:
			panic(fmt.Sprintf("invalid option type: %T", op))
		}
	}
	return opts
}

func (s *Signer) _sign(data string, opts *verifyOpts) string {
	mac := hmac.New(sha256.New, []byte(s.Key))
	io.WriteString(mac, data)
	if opts.writeNonce == nil {
		panic("Please provide WithExpiryTimestamp() or WithNonce()")
	}
	opts.writeNonce(mac)
	result := hex.EncodeToString(mac.Sum(nil))
	return result
}

// Verify checks data against expiryTs and sig. It validates using the provided options.
// You must provide either WithExpiryTimestamp(ts) or WithNonce(nonce).
func (s *Signer) Verify(data, sig string, ops ...any) (bool, error) {
	opts := s._parseOps(ops)
	if opts.err != nil {
		return false, opts.err
	}
	if err := opts.verifyFunc(); err != nil {
		return false, err
	}

	want := s._sign(data, opts)

	if !hmac.Equal([]byte(want), []byte(sig)) {
		return false, fmt.Errorf("signature mismatch: got %s, want %s", sig, want)
	}
	return true, nil
}

type verifyOpts struct {
	verifyFunc func() error
	err        error
	writeNonce func(w io.Writer)
}

func WithExpiryTimestamp(tsStr string) func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.writeNonce = func(w io.Writer) {
			w.Write([]byte(":" + tsStr))
		}
		exp, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			o.err = fmt.Errorf("invalid timestamp: %v", err)
			return
		}
		o.verifyFunc = func() error {
			if exp != 0 && time.Now().Unix() > exp {
				return fmt.Errorf("expired: %d < %d", exp, time.Now().Unix())
			}
			return nil
		}
	}
}

func WithExpiryTime(t time.Time) func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.writeNonce = func(w io.Writer) {
			w.Write([]byte(fmt.Sprintf(":%d", t.Unix())))
		}
		unix := t.Unix()
		o.verifyFunc = func() error {
			if unix != 0 && time.Now().Unix() > unix {
				return fmt.Errorf("expired: %d < %d", unix, time.Now().Unix())
			}
			return nil
		}
	}
}

func WithExpiryTimeUnix(unix int64) func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.writeNonce = func(w io.Writer) {
			w.Write([]byte(fmt.Sprintf(":%d", unix)))
		}
		o.verifyFunc = func() error {
			if unix != 0 && time.Now().Unix() > unix {
				return fmt.Errorf("expired: %d < %d", unix, time.Now().Unix())
			}
			return nil
		}
	}
}

func WithOutNonce() func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.writeNonce = func(w io.Writer) {}
		o.verifyFunc = func() error { return nil }
	}
}

func WithNonce(nonce string) func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.writeNonce = func(w io.Writer) {
			w.Write([]byte(":" + nonce))
		}
		o.verifyFunc = func() error { return nil }
	}
}

// WithExpiry is an alias for WithExpiryTime
func WithExpiry(t time.Time) func(*verifyOpts) {
	return WithExpiryTime(t)
}
