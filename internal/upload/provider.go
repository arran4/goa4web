package upload

import "context"

// Provider handles storing images using different backends.
type Provider interface {
	// Check verifies the backend is reachable and writable.
	Check(ctx context.Context) error
	// Write stores data under name.
	Write(ctx context.Context, name string, data []byte) error
	// Read retrieves data stored under name.
	Read(ctx context.Context, name string) ([]byte, error)
}

// CacheProvider extends Provider with a Cleanup method used for thumbnail caches.
type CacheProvider interface {
	Provider
	// Cleanup removes old cache files until the total size is below limit.
	Cleanup(ctx context.Context, limit int64) error
}
