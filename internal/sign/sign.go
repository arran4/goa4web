package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SignOption is a marker interface for sign options
type SignOption interface {
	isSignOption()
}

// WithNonce signs with a nonce instead of timestamp
type WithNonce string

func (WithNonce) isSignOption() {}

// WithExpiry signs with an expiry timestamp
type WithExpiry time.Time

func (WithExpiry) isSignOption() {}

// WithHostname includes hostname in signature data
type WithHostname string

func (WithHostname) isSignOption() {}

// Sign creates an HMAC signature for data with the given key and options.
// By default creates signature without nonce or expiry (not recommended for security).
// Use WithNonce or WithExpiry to add temporal validation.
func Sign(data string, key string, opts ...SignOption) string {
	var nonce string
	var expiry time.Time
	var hostname string

	for _, opt := range opts {
		switch v := opt.(type) {
		case WithNonce:
			nonce = string(v)
		case WithExpiry:
			expiry = time.Time(v)
		case WithHostname:
			hostname = string(v)
		case *legacyExpiryTimestamp:
			// Legacy: parse timestamp string
			if ts, err := strconv.ParseInt(v.ts, 10, 64); err == nil {
				expiry = time.Unix(ts, 0)
			}
		case *noNonce:
			// Legacy: explicitly no nonce/expiry
		}
	}

	mac := hmac.New(sha256.New, []byte(key))

	if hostname != "" {
		io.WriteString(mac, hostname)
		io.WriteString(mac, ":")
	}

	io.WriteString(mac, data)

	if nonce != "" {
		io.WriteString(mac, ":"+nonce)
	} else if !expiry.IsZero() {
		io.WriteString(mac, ":"+strconv.FormatInt(expiry.Unix(), 10))
	}

	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks if the signature is valid for the data.
// Returns error if signature doesn't match or if expiry has passed.
func Verify(data string, sig string, key string, opts ...SignOption) error {
	expected := Sign(data, key, opts...)
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return errors.New("signature mismatch")
	}

	// Check expiry
	for _, opt := range opts {
		switch v := opt.(type) {
		case WithExpiry:
			if time.Now().After(time.Time(v)) {
				return fmt.Errorf("signature expired at %v", time.Time(v))
			}
		case *legacyExpiryTimestamp:
			if ts, err := strconv.ParseInt(v.ts, 10, 64); err == nil {
				expiry := time.Unix(ts, 0)
				if time.Now().After(expiry) {
					return fmt.Errorf("signature expired at %v", expiry)
				}
			}
		}
	}

	return nil
}

// AddQuerySig adds signature and auth parameters as query parameters.
// If urlStr already has query params, they are preserved.
func AddQuerySig(urlStr string, sig string, opts ...SignOption) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("sig", sig)

	for _, opt := range opts {
		switch v := opt.(type) {
		case WithNonce:
			q.Set("nonce", string(v))
		case WithExpiry:
			q.Set("ts", strconv.FormatInt(time.Time(v).Unix(), 10))
		}
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// AddPathSig adds signature and auth parameters to the URL path.
// Format: /path/nonce/{nonce}/sign/{sig} or /path/ts/{ts}/sign/{sig}
func AddPathSig(urlStr string, sig string, opts ...SignOption) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	var authPart string
	for _, opt := range opts {
		switch v := opt.(type) {
		case WithNonce:
			authPart = fmt.Sprintf("/nonce/%s/sign/%s", url.PathEscape(string(v)), sig)
			break
		case WithExpiry:
			authPart = fmt.Sprintf("/ts/%d/sign/%s", time.Time(v).Unix(), sig)
			break
		}
	}

	if authPart == "" {
		authPart = "/sign/" + sig
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + authPart
	return u.String(), nil
}

// ExtractQuerySig extracts signature and auth options from query parameters.
// Returns sig and reconstructed options. The returned URL has sig/nonce/ts removed.
func ExtractQuerySig(urlStr string) (cleanURL string, sig string, opts []SignOption, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	sig = q.Get("sig")
	nonce := q.Get("nonce")
	tsStr := q.Get("ts")

	if nonce != "" {
		opts = append(opts, WithNonce(nonce))
	} else if tsStr != "" {
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			return "", "", nil, fmt.Errorf("invalid timestamp: %w", err)
		}
		opts = append(opts, WithExpiry(time.Unix(ts, 0)))
	}

	// Remove auth params
	q.Del("sig")
	q.Del("nonce")
	q.Del("ts")

	u.RawQuery = q.Encode()
	return u.String(), sig, opts, nil
}

// ExtractPathSig extracts signature from path-based auth.
// Looks for patterns like /nonce/{nonce}/sign/{sig} or /ts/{ts}/sign/{sig}
// Returns the clean path (without auth part), sig, and options.
func ExtractPathSig(path string, pathVars map[string]string) (cleanPath string, sig string, opts []SignOption, err error) {
	sig = pathVars["sign"]
	if sig == "" {
		sig = pathVars["sig"]
	}

	nonce := pathVars["nonce"]
	tsStr := pathVars["ts"]

	if nonce != "" {
		opts = append(opts, WithNonce(nonce))
		// Remove /nonce/{nonce}/sign/{sig} from path
		path = strings.TrimSuffix(path, fmt.Sprintf("/nonce/%s/sign/%s", nonce, sig))
	} else if tsStr != "" {
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			return "", "", nil, fmt.Errorf("invalid timestamp: %w", err)
		}
		opts = append(opts, WithExpiry(time.Unix(ts, 0)))
		// Remove /ts/{ts}/sign/{sig} from path
		path = strings.TrimSuffix(path, fmt.Sprintf("/ts/%s/sign/%s", tsStr, sig))
	} else {
		// Just /sign/{sig}
		path = strings.TrimSuffix(path, fmt.Sprintf("/sign/%s", sig))
	}

	return path, sig, opts, nil
}
