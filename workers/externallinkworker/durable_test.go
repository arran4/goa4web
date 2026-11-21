package externallinkworker

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
)

func TestDurableMetadataRefresh(t *testing.T) {
	const linkURL = "https://example.test/article" // The canonical target of the queued job.
	const imageURL = "https://example.test/image"  // The metadata image supplied by the fetch adapter.
	failure := errors.New("test dependency failure")
	for _, tc := range []struct {
		name          string
		recordURL     string
		hasTitle      bool
		lookupErr     error
		fetchErr      error
		downloadErr   error
		writeErr      error
		referenceErr  error
		wantErr       error
		fetches       int
		writes        int
		referenceRuns int
	}{
		{name: "writes bounded metadata and image reference", fetches: 1, writes: 1, referenceRuns: 1},
		{name: "populated link", hasTitle: true},
		{name: "removed record", lookupErr: sql.ErrNoRows},
		{name: "mismatched target", recordURL: "https://example.test/other"},
		{name: "download failure is optional", downloadErr: failure, fetches: 1, writes: 1},
		{name: "failed metadata write skips reference", writeErr: failure, wantErr: failure, fetches: 1, writes: 1},
		{name: "failed reference write reaches queue", referenceErr: failure, wantErr: failure, fetches: 1, writes: 1, referenceRuns: 1},
		{name: "fetch failure reaches queue without URL", fetchErr: errors.New(linkURL), fetches: 1},
		{name: "cancellation reaches queue", fetchErr: context.Canceled, wantErr: context.Canceled, fetches: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recordURL := tc.recordURL
			if recordURL == "" {
				recordURL = linkURL
			}
			q := &db.QuerierStub{
				GetExternalLinkByIDReturns:      &db.ExternalLink{ID: 42, Url: recordURL, CardTitle: sql.NullString{String: "title", Valid: tc.hasTitle}},
				GetExternalLinkByIDErr:          tc.lookupErr,
				UpdateExternalLinkMetadataErr:   tc.writeErr,
				UpdateExternalLinkImageCacheErr: tc.referenceErr,
			}
			var fetches int
			err := processExternalLink(t.Context(), q, 42, linkURL, func(_ context.Context, target string) (*opengraph.Info, error) {
				fetches++
				if target != linkURL {
					t.Fatalf("fetch target = %q", target)
				}
				if tc.fetchErr != nil {
					return nil, tc.fetchErr
				}
				return &opengraph.Info{Title: "Fetched", Image: imageURL, Description: strings.Repeat("🐚", 17000)}, nil
			}, func(_ context.Context, target string) (string, error) {
				if target != imageURL {
					t.Fatalf("cache target = %q", target)
				}
				return "cached", tc.downloadErr
			})
			wantErr := tc.wantErr
			if tc.fetchErr != nil {
				wantErr = tc.fetchErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("refresh error = %v, want %v", err, wantErr)
			}
			if err != nil && strings.Contains(err.Error(), linkURL) {
				t.Fatal("queued error exposes URL")
			}
			if fetches != tc.fetches || len(q.UpdateExternalLinkMetadataCalls) != tc.writes || len(q.UpdateExternalLinkImageCacheCalls) != tc.referenceRuns {
				t.Fatalf("fetches/writes/references = %d/%d/%d, want %d/%d/%d", fetches, len(q.UpdateExternalLinkMetadataCalls), len(q.UpdateExternalLinkImageCacheCalls), tc.fetches, tc.writes, tc.referenceRuns)
			}
			for _, stored := range q.UpdateExternalLinkMetadataCalls {
				if stored.ID != 42 || stored.CardTitle.String != "Fetched" || stored.CardImage.String != imageURL ||
					len(stored.CardDescription.String) != 64000 || !utf8.ValidString(stored.CardDescription.String) {
					t.Fatal("unexpected persisted metadata")
				}
			}
		})
	}
}

func TestDurableMetadataRetryRepairsMissingImageReference(t *testing.T) {
	const target = "https://example.test/article" // Target of both attempts.
	const image = "https://example.test/image"    // Image for the persisted card.
	record := &db.ExternalLink{ID: 42, Url: target}
	failure := errors.New("temporary reference write failure")
	q := &db.QuerierStub{
		GetExternalLinkByIDReturns: record,
		UpdateExternalLinkMetadataFn: func(_ context.Context, p db.UpdateExternalLinkMetadataParams) error {
			record.CardTitle = p.CardTitle
			record.CardImage = p.CardImage
			return nil
		},
		UpdateExternalLinkImageCacheErr: failure,
	}
	fetches := 0
	fetch := func(context.Context, string) (*opengraph.Info, error) {
		fetches++
		return &opengraph.Info{Title: "Fetched", Image: image}, nil
	}
	cache := func(context.Context, string) (string, error) { return "cached", nil }
	if err := processExternalLink(t.Context(), q, record.ID, target, fetch, cache); !errors.Is(err, failure) {
		t.Fatalf("first attempt error = %v, want %v", err, failure)
	}
	q.UpdateExternalLinkImageCacheErr = nil
	q.UpdateExternalLinkImageCacheFn = func(_ context.Context, p db.UpdateExternalLinkImageCacheParams) error {
		record.CardImageCache = p.CardImageCache
		return nil
	}
	if err := processExternalLink(t.Context(), q, record.ID, target, fetch, cache); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if fetches != 1 || !record.CardImageCache.Valid || record.CardImageCache.String != "cached" {
		t.Fatalf("retry fetched %d times and stored %v", fetches, record.CardImageCache)
	}
}
