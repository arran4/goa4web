package externallink

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/opengraph"
)

func TestRefreshMissingMetadata(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	cacheErr := errors.New("cache failed")
	storeErr := errors.New("store failed")
	for _, tc := range []struct {
		name                    string
		present                 bool
		fetchErr, cacheErr      error
		lookupErr, storeErr     error
		wantErr                 error
		attempts, updates, refs int
	}{
		{name: "already populated", present: true},
		{name: "lookup failed", lookupErr: storeErr, wantErr: storeErr},
		{name: "retry then cache and persist", attempts: 2, updates: 1, refs: 1},
		{name: "all fetches fail", fetchErr: fetchErr, wantErr: fetchErr, attempts: 3},
		{name: "cache failure still persists metadata", cacheErr: cacheErr, wantErr: cacheErr, attempts: 2, updates: 1},
		{name: "metadata write failure still records the cached image", storeErr: storeErr, wantErr: storeErr, attempts: 2, updates: 1, refs: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const linkID = 42                          // Link with an existing database identity.
			const url = "https://example.test/page"    // Canonical URL handed to the fetch adapter.
			const image = "https://example.test/image" // Image supplied by the fetched metadata.
			desc := strings.Repeat("🐚", 17000)
			info := &opengraph.Info{Title: "Page", Description: desc, Image: image, Author: "Author"}
			store := &metadataStoreStub{
				link:      &db.ExternalLink{ID: linkID, CardTitle: sql.NullString{String: "Present", Valid: tc.present}},
				lookupErr: tc.lookupErr,
				updateErr: tc.storeErr,
			}
			ctx := t.Context()
			var attempts, caches int
			fetch := func(ctx context.Context, target string) (*opengraph.Info, error) {
				return FetchMetadata(ctx, target, func(gotCtx context.Context, gotTarget string) (*opengraph.Info, error) {
					attempts++
					if gotCtx != ctx || gotTarget != url {
						t.Errorf("fetch context and target = %v, %q", gotCtx, gotTarget)
					}
					if tc.fetchErr != nil || attempts == 1 {
						return nil, fetchErr
					}
					return info, nil
				}, func(context.Context, time.Duration) error { return nil })
			}
			refreshed, err := RefreshMissingMetadata(ctx, store, linkID, url, fetch, func(_ context.Context, gotImage string) (string, error) {
				caches++
				if gotImage != image {
					t.Errorf("cache received %q, want %q", gotImage, image)
				}
				if tc.cacheErr != nil {
					return "", tc.cacheErr
				}
				return "cached-image", nil
			})
			if !errors.Is(err, tc.wantErr) || refreshed != (tc.updates == 1 && tc.storeErr == nil) {
				t.Errorf("refresh = (%v, %v), want (%v, %v)", refreshed, err, tc.updates == 1 && tc.storeErr == nil, tc.wantErr)
			}
			if attempts != tc.attempts || caches != tc.updates || len(store.updates) != tc.updates || len(store.refs) != tc.refs {
				t.Fatalf("attempts/cache calls/updates/image refs = %d/%d/%d/%d, want %d/%d/%d/%d",
					attempts, caches, len(store.updates), len(store.refs), tc.attempts, tc.updates, tc.updates, tc.refs)
			}
			if tc.updates > 0 {
				got := store.updates[0]
				if got.ID != linkID || got.CardTitle.String != "Page" || got.CardImage.String != image || got.CardAuthor.String != "Author" {
					t.Errorf("unexpected persisted metadata: %+v", got)
				}
				truncated := got.CardDescription.String
				if !got.CardDescription.Valid || !utf8.ValidString(truncated) || len(truncated) != 64000 || !strings.HasPrefix(desc, truncated) {
					t.Errorf("incorrect persisted description of %d bytes", len(truncated))
				}
				if info.Description != desc {
					t.Error("refresh changed the fetch adapter's metadata")
				}
			}
			if tc.refs > 0 && (store.refs[0].ID != linkID || store.refs[0].CardImageCache.String != "cached-image") {
				t.Errorf("incorrect persisted image cache reference: %+v", store.refs[0])
			}
		})
	}
}

type metadataStoreStub struct {
	link                 *db.ExternalLink
	lookupErr, updateErr error
	updates              []db.UpdateExternalLinkMetadataParams
	refs                 []db.UpdateExternalLinkImageCacheParams
}

func (s *metadataStoreStub) GetExternalLinkByID(context.Context, int32) (*db.ExternalLink, error) {
	return s.link, s.lookupErr
}

func (s *metadataStoreStub) UpdateExternalLinkMetadata(_ context.Context, p db.UpdateExternalLinkMetadataParams) error {
	s.updates = append(s.updates, p)
	return s.updateErr
}

func (s *metadataStoreStub) UpdateExternalLinkImageCache(_ context.Context, p db.UpdateExternalLinkImageCacheParams) error {
	s.refs = append(s.refs, p)
	return nil
}
