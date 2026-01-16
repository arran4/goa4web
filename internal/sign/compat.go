package sign

import (
	"time"
)

// Legacy compatibility functions for old code that hasn't been migrated yet

// WithExpiryTimestamp creates an expiry option from a timestamp string.
// Deprecated: Use WithExpiry with time.Time directly.
func WithExpiryTimestamp(tsStr string) SignOption {
	// Parse the timestamp - this is a legacy compatibility function
	// In the old code, timestamps were strings. We need to parse them.
	// For simplicity, if it doesn't parse, return a nonce option
	// This is for backward compatibility only
	return &legacyExpiryTimestamp{ts: tsStr}
}

type legacyExpiryTimestamp struct {
	ts string
}

func (l *legacyExpiryTimestamp) isSignOption() {}

// WithExpiryTime creates an expiry option from time.Time
func WithExpiryTime(t time.Time) SignOption {
	return WithExpiry(t)
}

// WithExpiryTimeUnix creates an expiry option from Unix timestamp
func WithExpiryTimeUnix(unix int64) SignOption {
	return WithExpiry(time.Unix(unix, 0))
}

// WithOutNonce creates a no-nonce option (not recommended for security).
// Deprecated: Signatures should always use nonce or expiry.
func WithOutNonce() SignOption {
	return &noNonce{}
}

type noNonce struct{}

func (noNonce) isSignOption() {}
