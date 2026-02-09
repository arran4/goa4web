package sign

import (
	"time"
)

// Legacy compatibility functions for old code that hasn't been migrated yet


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
