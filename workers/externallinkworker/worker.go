package externallinkworker

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/opengraph"
)

// Worker listens for new content and fetches metadata for external links.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier, cfg *config.RuntimeConfig) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)

	for {
		select {
		case msg := <-ch:
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				continue
			}
			if evt.Outcome != eventbus.TaskOutcomeSuccess {
				continue
			}
			body, ok := evt.Data["Body"].(string)
			if !ok || body == "" {
				continue
			}

			root, err := a4code.ParseString(body)
			if err != nil {
				continue
			}

			_ = a4code.Walk(root, func(n a4code.Node) error {
				link, ok := n.(*a4code.Link)
				if !ok {
					return nil
				}
				url := link.Href
				if url == "" {
					return nil
				}

				res, err := q.EnsureExternalLink(ctx, url)
				if err != nil {
					log.Printf("EnsureExternalLink %s: %v", url, err)
					return nil
				}
				id, err := res.LastInsertId()
				if err != nil {
					log.Printf("LastInsertId for %s: %v", url, err)
					return nil
				}

				// Check if we need to fetch metadata
				existing, err := q.GetExternalLinkByID(ctx, int32(id))
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					log.Printf("GetExternalLinkByID %d: %v", id, err)
					return nil
				}
				if existing != nil && existing.CardTitle.Valid && existing.CardTitle.String != "" {
					return nil // Already has title
				}

				// Use a safe client for fetching if available, but opengraph.Fetch creates one if nil.
				// However, to reuse client or settings, we can create one.
				// opengraph.Fetch uses its own client logic if nil is passed.

				title, desc, imageURL, err := opengraph.Fetch(url, nil)
				if err != nil {
					log.Printf("opengraph.Fetch %s: %v", url, err)
					return nil
				}

				var cachedImgName string
				if imageURL != "" {
					// Download and cache image
					// Use a simple client with timeout
					client := &http.Client{Timeout: 30 * time.Second}
					resp, err := client.Get(imageURL)
					if err == nil {
						defer resp.Body.Close()
						body, _ := io.ReadAll(resp.Body)
						if len(body) > 0 {
							im, _, err := image.Decode(bytes.NewReader(body))
							if err == nil {
								hash := fmt.Sprintf("%x", sha1.Sum(body))
								ext, err := intimages.CleanExtension(path.Ext(imageURL))
								if err == nil {
									// Store system image
									// We need CoreData to call StoreSystemImage
									cd := common.NewCoreData(ctx, q, cfg)
									name, err := cd.StoreSystemImage(common.StoreImageParams{
										ID:         hash,
										Ext:        ext,
										Data:       body,
										Image:      im,
										UploaderID: 0, // System
									})
									if err == nil {
										cachedImgName = name
									} else {
										log.Printf("StoreSystemImage error: %v", err)
									}
								}
							}
						}
					} else {
						log.Printf("Failed to download image %s: %v", imageURL, err)
					}
				}

				if err := q.UpdateExternalLinkMetadata(ctx, db.UpdateExternalLinkMetadataParams{
					CardTitle:       sql.NullString{String: title, Valid: title != ""},
					CardDescription: sql.NullString{String: desc, Valid: desc != ""},
					CardImage:       sql.NullString{String: imageURL, Valid: imageURL != ""},
					CardImageCache:  sql.NullString{String: cachedImgName, Valid: cachedImgName != ""},
					ID:              int32(id),
				}); err != nil {
					log.Printf("UpdateExternalLinkMetadata %d: %v", id, err)
				}

				return nil
			})

		case <-ctx.Done():
			return
		}
	}
}
