package api

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"golang.org/x/net/html"
)

type MetadataTask struct{ tasks.TaskString }

var metadataTask = &MetadataTask{TaskString: "api:metadata"}

var _ tasks.Task = (*MetadataTask)(nil)

type MetadataResponse struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageRef    string `json:"image_ref,omitempty"`
}

func (MetadataTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		return fmt.Errorf("url required %w", handlers.ErrBadRequest)
	}

	// Validate URL
	u, err := url.Parse(targetURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("invalid url %w", handlers.ErrBadRequest)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	// Create request with short timeout and limited reader
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return fmt.Errorf("request creation failed %w", err)
	}
	// Mimic a browser to avoid 403s
	req.Header.Set("User-Agent", "Goa4Web-Metadata-Fetcher/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Log error but return empty metadata is acceptable, or return error?
		// Returning error allows frontend to fallback.
		return fmt.Errorf("fetch failed %w", err)
	}
	defer resp.Body.Close()

	// Limit reading to 4K for metadata parsing as per requirements "truncated after 4k"
	// However, HTML head might be larger. Prompt said "truncated after 4k".
	// I'll read 4KB.
	limitReader := io.LimitReader(resp.Body, 4096)
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return fmt.Errorf("read failed %w", err)
	}

	meta := extractMetadata(bodyBytes)

	// Truncate title to 255 if too long (though requirement says truncate if "long enough",
	// frontend logic also handles "unbounded title" for Case 3.
	// But let's follow: "update / replace the link with the appropriate title if it is long enough, truncated to 255 if it's too long."
	// The frontend logic seems to imply it receives the full title and truncates if necessary for Case 1/2.
	// I will return the full title and let JS handle truncation logic, OR truncate here.
	// The prompt says "get the social media card ... then update ... with the appropriate title if it is long enough, truncated to 255 if it's too long."
	// This sounds like frontend logic.
	// But wait, prompt says "use write an a4code quote the link, then get the social media card ... via a new api call ... then update".
	// So I should return the title.

	if meta.ImageURL != "" {
		// If image URL is found, we need to download it and store it.
		// We need a new request for the image.
		// Note: The prompt says "gets the social media card image (if any) and adds it to the cache".
		// We need to do this safely.

		// Validate Image URL
		imgURL, err := url.Parse(meta.ImageURL)
		if err == nil {
			if imgURL.Scheme == "" {
				imgURL.Scheme = u.Scheme
				imgURL.Host = u.Host
			} else if strings.HasPrefix(meta.ImageURL, "//") {
				imgURL.Scheme = u.Scheme
			}
			// Download image
			imgRef, err := downloadAndStoreImage(ctx, cd, imgURL.String())
			if err == nil {
				meta.ImageRef = imgRef
			} else {
				// Ignore image download errors, just return metadata without image
				// log.Printf("Failed to download image: %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(meta); err != nil {
		return fmt.Errorf("json encode %w", handlers.ErrInternalServerError)
	}
	return nil
}

type extractedMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"-"`
	ImageRef    string `json:"image_ref,omitempty"`
}

func extractMetadata(data []byte) extractedMetadata {
	var meta extractedMetadata
	tokenizer := html.NewTokenizer(bytes.NewReader(data))

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		if tt == html.StartTagToken || tt == html.SelfClosingTagToken {
			t := tokenizer.Token()
			if t.Data == "meta" {
				var name, property, content string
				for _, attr := range t.Attr {
					if attr.Key == "name" {
						name = attr.Val
					}
					if attr.Key == "property" {
						property = attr.Val
					}
					if attr.Key == "content" {
						content = attr.Val
					}
				}

				switch property {
				case "og:title":
					meta.Title = content
				case "og:description":
					meta.Description = content
				case "og:image":
					meta.ImageURL = content
				}

				// Fallback to standard meta
				if name == "description" && meta.Description == "" {
					meta.Description = content
				}
				if name == "twitter:title" && meta.Title == "" {
					meta.Title = content
				}
				if name == "twitter:image" && meta.ImageURL == "" {
					meta.ImageURL = content
				}

			} else if t.Data == "title" {
				if tokenizer.Next() == html.TextToken {
					if meta.Title == "" {
						meta.Title = strings.TrimSpace(tokenizer.Token().Data)
					}
				}
			}
		}
	}
	return meta
}

func downloadAndStoreImage(ctx context.Context, cd *common.CoreData, imageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Goa4Web-Metadata-Fetcher/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	// Limit image size (e.g. 5MB)
	const maxImageSize = 5 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageSize))
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	// Generate ID
	id := fmt.Sprintf("%x", sha1.Sum(data))

	// Determine extension (simplified, maybe check Content-Type or magic numbers?)
	// common.StoreImage expects Ext.
	// Let's try to detect from response or assume based on format.
	// intimages.CleanExtension expects filename.
	// We can try to guess from URL or Content-Type.

	ext := "jpg" // default
	ct := resp.Header.Get("Content-Type")
	switch ct {
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	case "image/jpeg":
		ext = "jpg"
	case "image/webp":
		ext = "webp"
	default:
		// Try from URL
		if u, err := url.Parse(imageURL); err == nil {
			if e, err := intimages.CleanExtension(u.Path); err == nil && e != "" {
				ext = e
			}
		}
	}

	fname, err := cd.StoreImage(common.StoreImageParams{
		ID:         id,
		Ext:        ext,
		Data:       data,
		Image:      img,
		UploaderID: cd.UserID,
	})

	if err != nil {
		return "", err
	}

	if cd.ImageSigner != nil {
		return cd.ImageSigner.SignedRef("image:" + fname), nil
	}
	return "image:" + fname, nil
}
