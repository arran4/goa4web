package common

import (
	"context"
	"crypto/sha256"
	"fmt"
	_ "image/gif"  // Register format
	_ "image/jpeg" // Register format
	_ "image/png"  // Register format
	"io"
	"net/url"
	"os"
	"path"

	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/opengraph"
	"github.com/arran4/goa4web/internal/upload"
)

// DownloadAndCacheImage downloads an image from a URL, stores it using cd.StoreSystemImage,
// and returns the stored image name (cache key) prefixed with "image:".
func (cd *CoreData) DownloadAndCacheImage(imgURL string) (string, error) {
	client := opengraph.NewSafeClient() // Always use a safe client for external URLs

	resp, err := client.Get(imgURL)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("empty body")
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(body))
	// Try to get extension from URL path
	u, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	ext, err := intimages.CleanExtension(path.Base(u.Path))
	if err != nil {
		// Fallback to .jpg if no valid extension found in URL
		ext = ".jpg"
	}

	cacheRef := hash + ext
	sub1, sub2 := hash[:2], hash[2:4]
	key := path.Join(sub1, sub2, cacheRef)

	if cp := upload.CacheProviderFromConfig(cd.Config); cp != nil {
		if err := cp.Write(context.Background(), key, body); err != nil {
			return "", fmt.Errorf("cache write: %w", err)
		}
		// Optional: cleanup check? storeImageInternal does:
		// if ccp, ok := cp.(upload.CacheProvider); ok { ccp.Cleanup(...) }
		return "cache:" + cacheRef, nil
	}

	// Fallback if no cache provider? Or error?
	// serveCache supports local file system if no provider, using cfg.ImageCacheDir
	// We should write to ImageCacheDir if provider is nil?
	// upload.CacheProviderFromConfig usually returns a local provider if configured.
	// Let's check upload.CacheProviderFromConfig implementation or assume it handles local fallback if configured so.
	// Actually serveCache handles nil provider by using filepath directly.
	// We should probably replicate that or ensure we have a provider.
	// For now, let's assume we can rely on CacheProvider or manual write.

	// If CacheProviderFromConfig returns nil, we manually write to filesystem for local cache
	// mirroring serveCache logic.
	fullPath := path.Join(cd.Config.ImageCacheDir, sub1, sub2, cacheRef)
	if err := os.MkdirAll(path.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("mkdir cache: %w", err)
	}
	if err := os.WriteFile(fullPath, body, 0644); err != nil {
		return "", fmt.Errorf("write cache file: %w", err)
	}

	return "cache:" + cacheRef, nil
}
