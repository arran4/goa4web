package common

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"image"
	_ "image/gif"  // Register format
	_ "image/jpeg" // Register format
	_ "image/png"  // Register format
	"io"
	"net/url"
	"path"

	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/opengraph"
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

	im, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	hash := fmt.Sprintf("%x", sha1.Sum(body))
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

	// Store system image
	name, err := cd.StoreSystemImage(StoreImageParams{
		ID:         hash,
		Ext:        ext,
		Data:       body,
		Image:      im,
		UploaderID: 0, // System
	})
	if err != nil {
		return "", fmt.Errorf("store system image: %w", err)
	}

	return "image:" + name, nil
}
