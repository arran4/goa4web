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
// Use WithExpiry(time.Time) to set specific expiry.
// Use WithNonce(string) to use a nonce (valid indefinitely, assuming nonce is valid).
func (s *Signer) Sign(data string, ops ...any) (int64, string) {
	opts := &SignData{
		Expiry: time.Now().Add(24 * time.Hour),
	}
	for _, op := range ops {
		switch f := op.(type) {
		case func(*SignData):
			f(opts)
		default:
			panic(fmt.Sprintf("invalid option type: %T", op))
		}
	}

	mac := hmac.New(sha256.New, []byte(s.Key))
	if opts.Nonce != "" {
		io.WriteString(mac, fmt.Sprintf("%s:%s", data, opts.Nonce))
		return 0, hex.EncodeToString(mac.Sum(nil))
	}

	ts := opts.Expiry.Unix()
	io.WriteString(mac, fmt.Sprintf("%s:%d", data, ts))
	return ts, hex.EncodeToString(mac.Sum(nil))
}

type SignData struct {
	Expiry time.Time
	Nonce  string
}

func WithExpiry(t time.Time) func(*SignData) {
	return func(o *SignData) {
		o.Expiry = t
	}
}

func WithNonce(nonce string) func(*SignData) {
	return func(o *SignData) {
		o.Nonce = nonce
	}
}

// Verify checks data against expiryTs and sig. It validates using the provided options.
// You must provide either WithExpiryTimestamp(ts) or WithNonce(nonce).
func (s *Signer) Verify(data, sig string, ops ...any) (bool, error) {
	opts := &verifyOpts{}
	for _, op := range ops {
		switch f := op.(type) {
		case func(*verifyOpts):
			f(opts)
		default:
			panic(fmt.Sprintf("invalid option type: %T", op))
		}
	}
	if !opts.set {
		panic("Please provide WithExpiryTimestamp() or WithNonce()")
	}
	if opts.err != nil {
		return false, opts.err
	}
	if err := opts.verifyFunc(); err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(s.Key))
	if opts.nonce != "" {
		io.WriteString(mac, fmt.Sprintf("%s:%s", data, opts.nonce))
	} else {
		io.WriteString(mac, fmt.Sprintf("%s:%d", data, opts.expiryTs))
	}

	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(sig)) {
		return false, fmt.Errorf("signature mismatch: got %s, want %s", sig, want)
	}
	return true, nil
}

type verifyOpts struct {
	expiryTs   int64
	nonce      string
	verifyFunc func() error
	err        error
	set        bool
}

func WithExpiryTimestamp(tsStr string) func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.set = true
		exp, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			o.err = fmt.Errorf("invalid timestamp: %v", err)
			return
		}
		o.expiryTs = exp
		o.verifyFunc = func() error {
			if exp != 0 && time.Now().Unix() > exp {
				return fmt.Errorf("expired: %d < %d", exp, time.Now().Unix())
			}
			return nil
		}
	}
}

// Reuse usage of WithNonce for verification as well?
// The user might expect `WithNonce` to work for both if passed as `op ...any`.
// But `Sign` uses `*signOpts` and `Verify` uses `*verifyOpts`.
// I need `VerifyWithNonce` or overloading.
// User said `Replace WithoutExpiry() with WithNonce()`.
// This usually implies a single option name.
// But `Sign` and `Verify` take different structs.
// I can define `WithNonce` to return an interface or `any` that satisfies both?
// OR define two functions `WithNonce` and pass them?
// Or define `WithSignNonce` and `WithVerifyNonce`?
// Or better: `WithNonce` takes a `nonce` string.
// `Sign` looks for `func(*signOpts)`.
// `Verify` looks for `func(*verifyOpts)`.
// I can make `WithNonce` return a type that implements both?
// Or just define `WithNonce` to return `any`, but strict typing in `Sign`/`Verify` logic checks for `func ...`.
// I'll define `VerifyWithNonce` for clarity/safety OR `WithNonce` that returns a `VerifyOption`?
// The user snippet `sharesign.Sign(link, sign.WithoutExpiry())` used `WithoutExpiry`.
// I can define `WithNonce` to return `func(*signOpts)` AND `func(*verifyOpts)`? No, Go doesn't support intersection types like that easily.
// I will overload `WithNonce` to be a `SignOption`?
// And `WithVerifyNonce` for verify?
// Or simply copy `nonce` logic.
// Let's call the verify one `WithVerifyNonce` or `WithNonceVerifier`?
// Actually, I can use the same name `WithNonce` if I don't export both in the same package scope with conflict.
// But they are in the same package.
// I will name them explicitly to avoid confusion or use `VerifyWithNonce`.
// Wait, `WithExpiryTimestamp` is for Verify. `WithExpiry` is for Sign.
// So `WithNonce` for Sign. `WithVerifyNonce` for Verify?
// Or `WithNonce` for Sign, and `WithNonceVerification` for Verify?
// Let's use `WithVerifyNonce` for now to be safe.
// BUT `MakeImageURL` passes `ops` to `Sign`.
// `VerifyAndGetPath` calls `Verify`.
// They are separate call sites.

// NOTE: I will define `WithNonce` (for Sign) and `WithVerifyNonce` (for Verify) for now.
func WithVerifyNonce(nonce string) func(*verifyOpts) {
	return func(o *verifyOpts) {
		o.set = true
		o.nonce = nonce
		o.verifyFunc = func() error { return nil }
	}
}
