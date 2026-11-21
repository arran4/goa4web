package externallinkworker

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	domainexternallink "github.com/arran4/goa4web/internal/externallink"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/internal/opengraph"
)

// RunDurableWorker processes durable external_link jobs using lease fencing and cooldowns.
func RunDurableWorker(ctx context.Context, jobSvc *jobs.Service, q db.Querier, cfg *config.RuntimeConfig, workerID string, client *http.Client) {
	if jobSvc == nil || q == nil {
		return
	}
	jobSvc.RunWorkerLoop(ctx, workerID, 0, []string{jobs.JobTypeExternalLink}, func(ctx context.Context, job *db.DurableJob) error {
		if job == nil {
			return nil
		}
		return ProcessClaimedExternalLink(ctx, jobSvc, q, cfg, job, client)
	})
}

// ProcessClaimedExternalLink fetches outside the transaction and checks the
// durable claim when publishing external-link metadata or its image reference.
func ProcessClaimedExternalLink(ctx context.Context, svc *jobs.Service, q db.Querier, cfg *config.RuntimeConfig, job *db.DurableJob, client *http.Client) error {
	store := leasedExternalLinkStore{Querier: q, service: svc, job: job}
	return processExternalLink(ctx, store, int32(job.TargetID), job.TargetKey, func(ctx context.Context, target string) (*opengraph.Info, error) {
		return opengraph.Fetch(ctx, target, client)
	}, func(ctx context.Context, imageURL string) (string, error) {
		if err := svc.WriteWithActiveLease(ctx, job, func(db.Querier) error { return nil }); err != nil {
			return "", err
		}
		return common.NewCoreData(ctx, q, cfg, common.WithHTTPClient(client)).DownloadAndCacheImage(imageURL)
	})
}

type leasedExternalLinkStore struct {
	db.Querier
	service *jobs.Service
	job     *db.DurableJob
}

func (s leasedExternalLinkStore) EnsureExternalLink(ctx context.Context, target string) (sql.Result, error) {
	var result sql.Result
	err := s.service.WriteWithActiveLease(ctx, s.job, func(q db.Querier) error {
		var err error
		result, err = q.EnsureExternalLink(ctx, target)
		return err
	})
	return result, err
}

func (s leasedExternalLinkStore) UpdateExternalLinkMetadata(ctx context.Context, metadata db.UpdateExternalLinkMetadataParams) error {
	return s.service.WriteWithActiveLease(ctx, s.job, func(q db.Querier) error {
		return q.UpdateExternalLinkMetadata(ctx, metadata)
	})
}

func (s leasedExternalLinkStore) UpdateExternalLinkImageCache(ctx context.Context, reference db.UpdateExternalLinkImageCacheParams) error {
	return s.service.WriteWithActiveLease(ctx, s.job, func(q db.Querier) error {
		return q.UpdateExternalLinkImageCache(ctx, reference)
	})
}

func processExternalLink(ctx context.Context, q db.Querier, linkID int32, targetURL string,
	fetch func(context.Context, string) (*opengraph.Info, error),
	cacheImage func(context.Context, string) (string, error),
) error {
	canonicalURL := common.CanonicalizeExternalURL(targetURL)
	if canonicalURL == "" {
		return opengraph.ErrInvalidDestination
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
	if existing == nil || common.CanonicalizeExternalURL(existing.Url) != canonicalURL {
		return nil
	}
	cacheOptionalImage := func(ctx context.Context, imageURL string) (string, error) {
		ref, err := cacheImage(ctx, imageURL)
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if errors.Is(err, jobs.ErrLeaseLost) {
			return "", err
		}
		if err != nil {
			log.Printf("external link %d image cache unavailable", linkID)
			return "", nil
		}
		return ref, nil
	}
	if !domainexternallink.NeedsMetadata(existing) {
		return domainexternallink.RefreshMissingImageReference(ctx, q, existing, cacheOptionalImage)
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
	}, cacheOptionalImage)
	return errors.Join(err, cacheErr)
}

type metadataFetchError struct{ cause error }

func (e metadataFetchError) Error() string {
	var status *opengraph.HTTPStatusError
	if errors.As(e.cause, &status) {
		return status.Error()
	}
	if errors.Is(e.cause, opengraph.ErrInvalidDestination) {
		return opengraph.ErrInvalidDestination.Error()
	}
	return "external link fetch failed"
}
func (e metadataFetchError) Unwrap() error { return e.cause }
