package externallinkworker

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
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

				title, desc, image, err := opengraph.Fetch(url, nil)
				if err != nil {
					log.Printf("opengraph.Fetch %s: %v", url, err)
					return nil
				}

				var cachedImage string
				if image != "" {
					cd := common.NewCoreData(ctx, q, cfg)
					cached, err := cd.DownloadAndCacheImage(image)
					if err != nil {
						log.Printf("DownloadAndCacheImage %s: %v", image, err)
					} else {
						cachedImage = cached
					}
				}

				if err := q.UpdateExternalLinkMetadata(ctx, db.UpdateExternalLinkMetadataParams{
					CardTitle:       sql.NullString{String: title, Valid: title != ""},
					CardDescription: sql.NullString{String: desc, Valid: desc != ""},
					CardImage:       sql.NullString{String: image, Valid: image != ""},
					ID:              int32(id),
				}); err != nil {
					log.Printf("UpdateExternalLinkMetadata %d: %v", id, err)
				}

				if cachedImage != "" {
					if err := q.UpdateExternalLinkImageCache(ctx, db.UpdateExternalLinkImageCacheParams{
						CardImageCache: sql.NullString{String: cachedImage, Valid: true},
						ID:             int32(id),
					}); err != nil {
						log.Printf("UpdateExternalLinkImageCache %d: %v", id, err)
					}
				}

				return nil
			})

		case <-ctx.Done():
			return
		}
	}
}
