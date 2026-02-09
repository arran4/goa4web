package externallinkworker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/ast"
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
	ch, ack := bus.Subscribe(eventbus.TaskMessageType)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			func() {
				defer ack()
				evt, ok := msg.(eventbus.TaskEvent)
				if !ok {
					return
				}
				if evt.Outcome != eventbus.TaskOutcomeSuccess {
					return
				}
				body, ok := evt.Data["Body"].(string)
				if !ok || body == "" {
					return
				}

				root, err := a4code.ParseString(body)
				if err != nil {
					return
				}

				_ = ast.Walk(root, func(n ast.Node) error {
					link, ok := n.(*ast.Link)
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

					var info *opengraph.Info
					for i := 0; i < 3; i++ {
						info, err = opengraph.Fetch(url, nil)
						if err == nil {
							break
						}
						time.Sleep(time.Duration(i+1) * 2 * time.Second)
					}
					if err != nil {
						log.Printf("opengraph.Fetch %s: %v", url, err)
						return nil
					}

					var cachedImage string
					if info.Image != "" {
						cd := common.NewCoreData(ctx, q, cfg)
						cached, err := cd.DownloadAndCacheImage(info.Image)
						if err != nil {
							log.Printf("DownloadAndCacheImage %s: %v", info.Image, err)
						} else {
							cachedImage = cached
						}
					}

					if err := q.UpdateExternalLinkMetadata(ctx, db.UpdateExternalLinkMetadataParams{
						CardTitle:       sql.NullString{String: info.Title, Valid: info.Title != ""},
						CardDescription: sql.NullString{String: info.Description, Valid: info.Description != ""},
						CardImage:       sql.NullString{String: info.Image, Valid: info.Image != ""},
						CardDuration:    sql.NullString{String: info.Duration, Valid: info.Duration != ""},
						CardUploadDate:  sql.NullString{String: info.UploadDate, Valid: info.UploadDate != ""},
						CardAuthor:      sql.NullString{String: info.Author, Valid: info.Author != ""},
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
			}()

		case <-ctx.Done():
			return
		}
	}
}
