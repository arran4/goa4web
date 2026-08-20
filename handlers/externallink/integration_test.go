package externallink

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/arran4/goa4web/workers/externallinkworker"
	"github.com/stretchr/testify/assert"
)

func TestExternalLinkEndToEndIdentityAndPipeline(t *testing.T) {
	key := "test-secret-key"
	dbLinks := make(map[string]*db.ExternalLink)
	dbClicks := make(map[string]int32)
	var nextID int32 = 100

	qs := testhelpers.NewQuerierStub()
	qs.EnsureExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
		if l, ok := dbLinks[url]; ok {
			return db.FakeSQLResult{LastInsertIDValue: int64(l.ID)}, nil
		}
		id := nextID
		nextID++
		dbLinks[url] = &db.ExternalLink{
			ID:     id,
			Url:    url,
			Clicks: 0,
		}
		return db.FakeSQLResult{LastInsertIDValue: int64(id)}, nil
	}
	qs.GetExternalLinkFn = func(ctx context.Context, url string) (*db.ExternalLink, error) {
		if l, ok := dbLinks[url]; ok {
			return l, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.GetExternalLinkByIDFn = func(ctx context.Context, id int32) (*db.ExternalLink, error) {
		for _, l := range dbLinks {
			if l.ID == id {
				return l, nil
			}
		}
		return nil, sql.ErrNoRows
	}
	qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
		for _, l := range dbLinks {
			if l.ID == arg.ID {
				l.CardTitle = arg.CardTitle
				l.CardDescription = arg.CardDescription
				l.CardImage = arg.CardImage
				return nil
			}
		}
		return nil
	}
	qs.SystemRegisterExternalLinkClickFn = func(ctx context.Context, url string) error {
		dbClicks[url]++
		if l, ok := dbLinks[url]; ok {
			l.Clicks++
		}
		return nil
	}

	cfg := &config.RuntimeConfig{
		BaseURL: "http://example.org",
	}

	// 1. Worker parses post body containing tracked link
	rawTrackedURL := "https://example.com/blog/post?id=42&utm_source=twitter&utm_medium=social"
	canonicalURL := "https://example.com/blog/post?id=42"

	bus := eventbus.NewBus()
	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go externallinkworker.Worker(workerCtx, bus, qs, cfg)

	// Ensure worker has subscribed
	assert.Eventually(t, func() bool {
		_ = bus.Publish(eventbus.TaskEvent{
			Outcome: eventbus.TaskOutcomeSuccess,
			Data: map[string]any{
				"Body": "[link=" + rawTrackedURL + "]My Post[/link]",
			},
		})
		_, ok := dbLinks[canonicalURL]
		return ok
	}, 3e9, 5e7, "worker should insert link under canonical URL")

	// Verify worker stored under canonicalURL, not rawTrackedURL
	assert.Nil(t, dbLinks[rawTrackedURL], "worker must not store raw tracked URL")
	assert.NotNil(t, dbLinks[canonicalURL], "worker must store canonical URL")

	// Set metadata on canonical record
	dbLinks[canonicalURL].CardTitle = sql.NullString{String: "My Post Title", Valid: true}
	dbLinks[canonicalURL].CardDescription = sql.NullString{String: "My Post Description", Valid: true}

	// 2. Renderer renders the link from post content
	cd := common.NewCoreData(context.Background(), qs, cfg, common.WithLinkSignKey(key))
	provider := common.NewGoa4WebLinkProvider(cd, context.Background())

	open, close, _ := provider.RenderLink(rawTrackedURL, false, false)
	rendered := open + close

	// Assert rendered output routes through /goto with canonical URL and signature
	assert.Contains(t, rendered, "http://example.org/goto?u=https%3A%2F%2Fexample.com%2Fblog%2Fpost%3Fid%3D42&sig=")
	assert.NotContains(t, rendered, "utm_source", "rendered link must not expose tracking parameters")
	assert.Contains(t, rendered, "title=\"https://example.com/blog/post?id=42 - My Post Title - My Post Description\"", "tooltip must use canonical URL")

	// 3. User clicks link: hits RedirectHandler with /goto?u=canonicalURL&sig=...&go=1
	gotoURL := cd.SignLinkURL(rawTrackedURL) // SignLinkURL canonicalizes internally
	reqURL := strings.TrimPrefix(gotoURL, "http://example.org") + "&go=1"

	req := httptest.NewRequest("GET", reqURL, nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	RedirectHandler(rec, req)

	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, canonicalURL, rec.Header().Get("Location"), "redirect destination must be canonical URL")
	assert.Equal(t, int32(1), dbClicks[canonicalURL], "click must be registered under canonical URL")
	assert.Equal(t, int32(0), dbClicks[rawTrackedURL], "no click should be registered under raw tracked URL")
}

func TestExternalLinkPreservesBloombergAccessTokenEndToEnd(t *testing.T) {
	key := "test-secret-key"
	dbLinks := make(map[string]*db.ExternalLink)

	qs := testhelpers.NewQuerierStub()
	qs.GetExternalLinkFn = func(ctx context.Context, url string) (*db.ExternalLink, error) {
		if l, ok := dbLinks[url]; ok {
			return l, nil
		}
		return nil, sql.ErrNoRows
	}

	cfg := &config.RuntimeConfig{
		BaseURL: "http://example.org",
	}

	rawBloombergURL := "https://www.bloomberg.com/news/articles/2026-01-01/article?accessToken=secret_token_abc123&utm_source=twitter"
	canonicalURL := "https://www.bloomberg.com/news/articles/2026-01-01/article?accessToken=secret_token_abc123"

	dbLinks[canonicalURL] = &db.ExternalLink{
		ID:              200,
		Url:             canonicalURL,
		CardTitle:       sql.NullString{String: "Bloomberg Markets", Valid: true},
		CardDescription: sql.NullString{String: "Financial news analysis", Valid: true},
	}

	cd := common.NewCoreData(context.Background(), qs, cfg, common.WithLinkSignKey(key))
	provider := common.NewGoa4WebLinkProvider(cd, context.Background())

	open, close, _ := provider.RenderLink(rawBloombergURL, true, true)
	rendered := open + close

	assert.Contains(t, rendered, "accessToken=secret_token_abc123", "rendered card must preserve accessToken")
	assert.NotContains(t, rendered, "utm_source", "rendered card must remove utm_source")
	assert.Contains(t, rendered, "Bloomberg Markets")

	// Goto redirect preserves accessToken
	gotoURL := cd.SignLinkURL(rawBloombergURL)
	reqURL := strings.TrimPrefix(gotoURL, "http://example.org") + "&go=1"

	req := httptest.NewRequest("GET", reqURL, nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	RedirectHandler(rec, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, canonicalURL, rec.Header().Get("Location"), "redirect destination must preserve accessToken")
}

func TestMigratedExternalLinkLookupAgreement(t *testing.T) {
	key := "test-secret-key"
	dbLinks := make(map[string]*db.ExternalLink)

	qs := testhelpers.NewQuerierStub()
	qs.GetExternalLinkFn = func(ctx context.Context, url string) (*db.ExternalLink, error) {
		if l, ok := dbLinks[url]; ok {
			return l, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.EnsureExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
		if l, ok := dbLinks[url]; ok {
			return db.FakeSQLResult{LastInsertIDValue: int64(l.ID)}, nil
		}
		t.Fatalf("unexpected insertion of second identity: %s", url)
		return nil, nil
	}

	cfg := &config.RuntimeConfig{
		BaseURL: "http://example.org",
	}

	// Simulated row cleaned by migration 0096
	// Original raw tracked URL was: "https://example.com/item?id=1&utm_source=x&category=books"
	// Migration 0096 left: "https://example.com/item?id=1&category=books"
	migratedStoredURL := "https://example.com/item?id=1&category=books"
	dbLinks[migratedStoredURL] = &db.ExternalLink{
		ID:              300,
		Url:             migratedStoredURL,
		Clicks:          15,
		CardTitle:       sql.NullString{String: "Item 1", Valid: true},
		CardDescription: sql.NullString{String: "Item Description", Valid: true},
	}

	cd := common.NewCoreData(context.Background(), qs, cfg, common.WithLinkSignKey(key))

	// Application runtime receives the tracked URL in content
	rawTrackedURL := "https://example.com/item?id=1&utm_source=x&category=books"

	// 1. Runtime canonicalization matches migrated stored URL exactly
	runtimeCanonical := common.CanonicalizeExternalURL(rawTrackedURL)
	assert.Equal(t, migratedStoredURL, runtimeCanonical, "runtime canonical URL must exactly match migrated URL")

	// 2. cd.GetExternalLink finds the pre-existing migrated record
	link, err := cd.GetExternalLink(context.Background(), rawTrackedURL)
	assert.NoError(t, err)
	assert.NotNil(t, link)
	assert.Equal(t, int32(300), link.ID)
	assert.Equal(t, int32(15), link.Clicks)
	assert.Equal(t, "Item 1", link.CardTitle.String)

	// 3. Renderer renders using the existing metadata rather than fallback
	provider := common.NewGoa4WebLinkProvider(cd, context.Background())
	open, close, _ := provider.RenderLink(rawTrackedURL, true, true)
	rendered := open + close
	assert.Contains(t, rendered, "Item 1", "must render migrated card title")
	assert.Contains(t, rendered, "Item Description", "must render migrated card description")
	assert.Contains(t, rendered, "http://example.org/goto?u=https%3A%2F%2Fexample.com%2Fitem%3Fid%3D1%26category%3Dbooks&sig=")
}

// Compile-time check that a4code is imported and reachable
var _ = a4code.ParseString
