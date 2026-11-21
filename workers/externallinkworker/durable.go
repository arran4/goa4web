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
	return processExternalLink(ctx, q, linkID, targetURL, func(ctx context.Context, target string) (*opengraph.Info, error) {
		return domainexternallink.FetchMetadata(ctx, target, func(ctx context.Context, u string) (*opengraph.Info, error) {
			return opengraph.Fetch(ctx, u, nil)
		}, domainexternallink.Wait)
	}, func(ctx context.Context, imageURL string) (string, error) {
		return common.NewCoreData(ctx, q, cfg).DownloadAndCacheImage(imageURL)
	})
}

func processExternalLink(ctx context.Context, q db.Querier, linkID int32, targetURL string,
	fetch func(context.Context, string) (*opengraph.Info, error),
	cacheImage func(context.Context, string) (string, error),
) error {
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil // The record was removed after the job was queued.
	}
	if err != nil {
		return err
	}
	if existing == nil || common.CanonicalizeExternalURL(existing.Url) != canonicalURL || !domainexternallink.NeedsMetadata(existing) {
		return nil
	}
	cacheErr, err := domainexternallink.RefreshMetadata(ctx, q, canonicalURL, func(context.Context) (int32, error) {
		return linkID, ctx.Err()
	}, func(ctx context.Context, target string) (*opengraph.Info, error) {
		info, err := fetch(ctx, target)
		if err != nil {
			// The jobs service stores error strings and may send them to operators.
			return nil, metadataFetchError{cause: err}
		}
		return info, nil
	}, func(ctx context.Context, imageURL string) (string, error) {
		ref, err := cacheImage(ctx, imageURL)
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err != nil {
			log.Printf("external link %d image cache unavailable", linkID)
			return "", nil
		}
		return ref, err
	})
	return errors.Join(err, cacheErr)
}

type metadataFetchError struct{ cause error }

func (e metadataFetchError) Error() string { return "external link fetch failed" }
func (e metadataFetchError) Unwrap() error { return e.cause }
