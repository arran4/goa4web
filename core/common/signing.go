package common

import (
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
)

// SignShareURL signs a share URL with path-based signature by default.
// The path will have "/shared" injected after the module name.
func (cd *CoreData) SignShareURL(path string, opts ...sign.SignOption) (string, error) {
	sharedPath := signutil.InjectShared(path)

	// Add nonce if no options provided
	if len(opts) == 0 {
		opts = append(opts, sign.WithNonce(signutil.GenerateNonce()))
	}

	fullURL := strings.TrimSuffix(cd.Config.HTTPHostname, "/") + "/" + strings.TrimPrefix(sharedPath, "/")
	return signutil.SignAndAddPath(fullURL, sharedPath, cd.ShareSignKey, opts...)
}

// SignShareURLQuery is like SignShareURL but uses query-based signatures.
func (cd *CoreData) SignShareURLQuery(path string, opts ...sign.SignOption) (string, error) {
	sharedPath := signutil.InjectShared(path)

	// Add nonce if no options provided
	if len(opts) == 0 {
		opts = append(opts, sign.WithNonce(signutil.GenerateNonce()))
	}

	fullURL := strings.TrimSuffix(cd.Config.HTTPHostname, "/") + "/" + strings.TrimPrefix(sharedPath, "/")
	return signutil.SignAndAddQuery(fullURL, sharedPath, cd.ShareSignKey, opts...)
}

// SignImageURL signs an image URL with the given TTL.
// Returns a signed URL that can be used to access the image.
func (cd *CoreData) SignImageURL(imageRef string, ttl time.Duration) string {
	// Strip image: or img: prefix if present
	imageRef = strings.TrimPrefix(strings.TrimPrefix(imageRef, "image:"), "img:")

	path := "/images/image/" + imageRef
	expiry := time.Now().Add(ttl)

	fullURL := strings.TrimSuffix(cd.Config.HTTPHostname, "/") + path
	signedURL, _ := signutil.SignAndAddQuery(fullURL, path, cd.ImageSignKey, sign.WithExpiry(expiry))
	return signedURL
}

// SignCacheURL signs a cache URL with the given TTL.
func (cd *CoreData) SignCacheURL(cacheRef string, ttl time.Duration) string {
	path := "/images/cache/" + cacheRef
	expiry := time.Now().Add(ttl)

	fullURL := strings.TrimSuffix(cd.Config.HTTPHostname, "/") + path
	signedURL, _ := signutil.SignAndAddQuery(fullURL, path, cd.ImageSignKey, sign.WithExpiry(expiry))
	return signedURL
}

// SignLinkURL signs an external link redirect URL.
func (cd *CoreData) SignLinkURL(externalURL string) string {
	data := "link:" + externalURL
	sig := sign.Sign(data, cd.LinkSignKey, sign.WithOutNonce())

	// Return /goto?u={url}&sig={sig}
	return strings.TrimSuffix(cd.Config.HTTPHostname, "/") + "/goto?u=" + externalURL + "&sig=" + sig
}

// SignFeedURL signs a feed URL for authenticated access.
// Format: /{section}/u/{username}/{rest}?sig={sig}
func (cd *CoreData) SignFeedURL(path, username string) string {
	data := "feed:" + username + ":" + path
	sig := sign.Sign(data, cd.FeedSignKey, sign.WithOutNonce())

	// Inject /u/{username} after first segment
	parts := strings.SplitN(path, "/", 3)
	var newPath string
	if len(parts) >= 3 && parts[0] == "" && parts[1] != "" {
		newPath = "/" + parts[1] + "/u/" + username
		if len(parts) > 2 {
			newPath += "/" + parts[2]
		}
	} else {
		newPath = "/u/" + username + path
	}

	return strings.TrimSuffix(cd.Config.HTTPHostname, "/") + "/" + strings.TrimPrefix(newPath, "/") + "?sig=" + sig
}

// MapImageURL converts image references to signed URLs.
// Used by a4code mapper.
func (cd *CoreData) MapImageURL(tag, val string) string {
	if tag != "img" {
		return val
	}

	switch {
	case strings.HasPrefix(val, "uploading:"):
		return val
	case strings.HasPrefix(val, "image:") || strings.HasPrefix(val, "img:"):
		return cd.SignImageURL(val, 24*time.Hour)
	case strings.HasPrefix(val, "cache:"):
		cacheRef := strings.TrimPrefix(val, "cache:")
		return cd.SignCacheURL(cacheRef, 24*time.Hour)
	default:
		return val
	}
}

// MapLinkURL converts external links to signed redirect URLs.
// Used by a4code mapper.
func (cd *CoreData) MapLinkURL(tag, val string) string {
	if tag != "a" {
		return val
	}

	// Only sign external links (http:// or https://)
	if !strings.HasPrefix(val, "http://") && !strings.HasPrefix(val, "https://") {
		return val
	}

	// Check if it's an allowed host
	// For now, always sign external links
	// TODO: Add hostname checking if needed

	return cd.SignLinkURL(val)
}
