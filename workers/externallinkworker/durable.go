package externallinkworker

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	domainexternallink "github.com/arran4/goa4web/internal/externallink"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
)

// RunDurableWorker processes durable external_link jobs using lease fencing and cooldowns.
func RunDurableWorker(ctx context.Context, jobSvc *jobs.Service, q db.Querier, cfg *config.RuntimeConfig, workerID string) {
	if jobSvc == nil || q == nil {
		return
	}
	jobSvc.RunWorkerLoop(ctx, workerID, 0, []string{jobs.JobTypeExternalLink}, func(ctx context.Context, job *db.DurableJob) error {
		if job == nil {
			return nil
		}
		return ProcessExternalLink(ctx, q, cfg, int32(job.TargetID), job.TargetKey)
	})
}

// ProcessExternalLink resolves, fetches OpenGraph metadata for, and persists an external link.
func ProcessExternalLink(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig, linkID int32, targetURL string) error {
	canonicalURL := common.CanonicalizeExternalURL(targetURL)
	if canonicalURL == "" {
		return nil
	}
	if linkID == 0 {
		resolvedID, err := domainexternallink.ResolveRecord(ctx, q, canonicalURL)
		if err != nil {
			return err
		}
		linkID = resolvedID
	}
	if linkID == 0 {
		return nil
	}
	existing, err := q.GetExternalLinkByID(ctx, linkID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if !domainexternallink.NeedsMetadata(existing) {
		return nil
	}
	info, err := domainexternallink.FetchMetadata(ctx, canonicalURL, func(ctx context.Context, u string) (*opengraph.Info, error) {
		return opengraph.Fetch(ctx, u, nil)
	}, domainexternallink.Wait)
	if err != nil {
		return err
	}
	if info == nil {
		return nil
	}
	var cachedImage string
	if info.Image != "" {
		cd := common.NewCoreData(ctx, q, cfg)
		if cached, cacheErr := cd.DownloadAndCacheImage(info.Image); cacheErr != nil {
			log.Printf("DownloadAndCacheImage %s: %v", info.Image, cacheErr)
		} else {
			cachedImage = cached
		}
	}
	desc := info.Description
	if len(desc) > 65000 {
		runes := []rune(desc)
		if len(runes) > 16000 {
			desc = string(runes[:16000])
		}
		for len(desc) > 65000 {
			runes = []rune(desc)
			desc = string(runes[:len(runes)-100])
		}
	}
	if err := q.UpdateExternalLinkMetadata(ctx, db.UpdateExternalLinkMetadataParams{
		CardTitle:       sql.NullString{String: info.Title, Valid: info.Title != ""},
		CardDescription: sql.NullString{String: desc, Valid: desc != ""},
		CardImage:       sql.NullString{String: info.Image, Valid: info.Image != ""},
		CardDuration:    sql.NullString{String: info.Duration, Valid: info.Duration != ""},
		CardUploadDate:  sql.NullString{String: info.UploadDate, Valid: info.UploadDate != ""},
		CardAuthor:      sql.NullString{String: info.Author, Valid: info.Author != ""},
		ID:              linkID,
	}); err != nil {
		return err
	}
	if cachedImage != "" {
		if err := q.UpdateExternalLinkImageCache(ctx, db.UpdateExternalLinkImageCacheParams{
			CardImageCache: sql.NullString{String: cachedImage, Valid: true},
			ID:             linkID,
		}); err != nil {
			return err
		}
	}
	return nil
}
