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

func TestExternalLinkUppercaseSchemeRendererEndToEnd(t *testing.T) {
	key := "test-secret-key"
	dbLinks := make(map[string]*db.ExternalLink)

	rawTrackedURL := "HTTPS://example.com/article?id=1&utm_source=x"
	canonicalURL := "HTTPS://example.com/article?id=1"

	dbLinks[canonicalURL] = &db.ExternalLink{
		ID:              400,
		Url:             canonicalURL,
		CardTitle:       sql.NullString{String: "Uppercase Scheme Article", Valid: true},
		CardDescription: sql.NullString{String: "Preserves uppercase scheme", Valid: true},
	}

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

	cd := common.NewCoreData(context.Background(), qs, cfg, common.WithLinkSignKey(key))
	provider := common.NewGoa4WebLinkProvider(cd, context.Background())

	// 1. Verify canonical URL preserves uppercase scheme
	assert.Equal(t, canonicalURL, common.CanonicalizeExternalURL(rawTrackedURL))

	// 2. Render link as card
	open, close, _ := provider.RenderLink(rawTrackedURL, true, true)
	rendered := open + close

	// Assert rendered output routes through /goto with canonical uppercase URL and signature
	assert.Contains(t, rendered, "http://example.org/goto?u=HTTPS%3A%2F%2Fexample.com%2Farticle%3Fid%3D1&sig=")
	assert.NotContains(t, rendered, "utm_source", "rendered link must not expose tracking parameters")
	assert.Contains(t, rendered, "Uppercase Scheme Article")
	assert.Contains(t, rendered, "HTTPS://example.com/article?id=1", "must contain uppercase canonical URL in card footer/title")
}

func TestMigratedExternalLinkLookupAgreement(t *testing.T) {
	key := "test-secret-key"

	cases := []struct {
		name              string
		rawTrackedURL     string
		migratedStoredURL string
		id                int32
		title             string
		desc              string
	}{
		{
			name:              "Standard tracking parameter removal",
			rawTrackedURL:     "https://example.com/item?id=1&utm_source=x&category=books",
			migratedStoredURL: "https://example.com/item?id=1&category=books",
			id:                300,
			title:             "Item 1",
			desc:              "Item Description",
		},
		{
			name:              "UTM prefix with hyphen removed",
			rawTrackedURL:     "https://example.com/utm1?utm_campaign-name=x&id=1",
			migratedStoredURL: "https://example.com/utm1?id=1",
			id:                312,
			title:             "UTM Hyphen",
			desc:              "UTM Hyphen Description",
		},
		{
			name:              "UTM prefix with dot removed",
			rawTrackedURL:     "https://example.com/utm2?utm_custom.value=x&id=1",
			migratedStoredURL: "https://example.com/utm2?id=1",
			id:                313,
			title:             "UTM Dot",
			desc:              "UTM Dot Description",
		},
		{
			name:              "UTM uppercase prefix with hyphen removed",
			rawTrackedURL:     "https://example.com/utm3?UTM_custom-value=x&id=1",
			migratedStoredURL: "https://example.com/utm3?id=1",
			id:                314,
			title:             "UTM Upper Hyphen",
			desc:              "UTM Upper Hyphen Description",
		},
		{
			name:              "Non-UTM prefix key utm.foo preserved",
			rawTrackedURL:     "https://example.com/utm4?utm.foo=x&id=1",
			migratedStoredURL: "https://example.com/utm4?utm.foo=x&id=1",
			id:                315,
			title:             "Non-UTM Key",
			desc:              "Non-UTM Key Description",
		},
		{
			name:              "Preserve exact-key prefix gclid_extra when tracking param is removed",
			rawTrackedURL:     "https://example.com/prefix1?id=1&gclid_extra=keep&utm_source=x",
			migratedStoredURL: "https://example.com/prefix1?id=1&gclid_extra=keep",
			id:                316,
			title:             "GCLID Extra",
			desc:              "GCLID Extra Description",
		},
		{
			name:              "Preserve exact-key prefix fbclid_extra when tracking param is removed",
			rawTrackedURL:     "https://example.com/prefix2?id=1&fbclid_extra=val&utm_medium=mail",
			migratedStoredURL: "https://example.com/prefix2?id=1&fbclid_extra=val",
			id:                317,
			title:             "FBCLID Extra",
			desc:              "FBCLID Extra Description",
		},
		{
			name:              "Preserve exact-key prefix clickidfoo when tracking param is removed",
			rawTrackedURL:     "https://example.com/prefix3?id=1&clickidfoo=123&gclid=real",
			migratedStoredURL: "https://example.com/prefix3?id=1&clickidfoo=123",
			id:                318,
			title:             "ClickID Foo",
			desc:              "ClickID Foo Description",
		},
		{
			name:              "Empty query components between params preserved",
			rawTrackedURL:     "https://example.com/search1?a=1&&utm_source=x&b=2",
			migratedStoredURL: "https://example.com/search1?a=1&&b=2",
			id:                301,
			title:             "Search 1",
			desc:              "Search Description 1",
		},
		{
			name:              "Leading empty query component preserved",
			rawTrackedURL:     "https://example.com/search2?&a=1&utm_source=x",
			migratedStoredURL: "https://example.com/search2?&a=1",
			id:                302,
			title:             "Search 2",
			desc:              "Search Description 2",
		},
		{
			name:              "Trailing empty query component preserved",
			rawTrackedURL:     "https://example.com/search3?a=1&utm_source=x&",
			migratedStoredURL: "https://example.com/search3?a=1&",
			id:                303,
			title:             "Search 3",
			desc:              "Search Description 3",
		},
		{
			name:              "Multiple empty query components preserved",
			rawTrackedURL:     "https://example.com/search4?a=1&&utm_source=x&&b=2",
			migratedStoredURL: "https://example.com/search4?a=1&&&b=2",
			id:                304,
			title:             "Search 4",
			desc:              "Search Description 4",
		},
		{
			name:              "Percent-encoded key treated conservatively as non-tracking in migration and runtime",
			rawTrackedURL:     "https://example.com/search5?%75tm_source=x&id=1",
			migratedStoredURL: "https://example.com/search5?%75tm_source=x&id=1",
			id:                305,
			title:             "Search 5",
			desc:              "Search Description 5",
		},
		{
			name:              "Uppercase scheme case preserved",
			rawTrackedURL:     "HTTPS://example.com/upper?id=1&utm_source=x",
			migratedStoredURL: "HTTPS://example.com/upper?id=1",
			id:                306,
			title:             "Item Upper",
			desc:              "Uppercase Scheme Description",
		},
		{
			name:              "Leading empty component preserved as bare question mark",
			rawTrackedURL:     "https://example.com/bare1?&utm_source=x",
			migratedStoredURL: "https://example.com/bare1?",
			id:                307,
			title:             "Search Bare Q",
			desc:              "Bare Question Mark",
		},
		{
			name:              "Trailing empty component preserved as bare question mark",
			rawTrackedURL:     "https://example.com/bare2?utm_source=x&",
			migratedStoredURL: "https://example.com/bare2?",
			id:                308,
			title:             "Search Bare Q Trailing",
			desc:              "Bare Question Mark Trailing",
		},
		{
			name:              "Leading and trailing empty components preserved as ?&",
			rawTrackedURL:     "https://example.com/bare3?&utm_source=x&",
			migratedStoredURL: "https://example.com/bare3?&",
			id:                309,
			title:             "Search Amp",
			desc:              "Ampersand remaining",
		},
		{
			name:              "Non-http schemes preserved untouched in migration and runtime",
			rawTrackedURL:     "mailto:user@example.com?utm_source=x",
			migratedStoredURL: "mailto:user@example.com?utm_source=x",
			id:                310,
			title:             "Mailto",
			desc:              "Mailto Description",
		},
		{
			name:              "Unusual host, userinfo, port, path escaping, and fragment preserved",
			rawTrackedURL:     "https://User:Pass@EXAMPLE.com:8080/path/to/page%201?id=42&utm_source=x#section-1",
			migratedStoredURL: "https://User:Pass@EXAMPLE.com:8080/path/to/page%201?id=42#section-1",
			id:                311,
			title:             "Unusual URL",
			desc:              "Unusual URL Description",
		},
	}

	dbLinks := make(map[string]*db.ExternalLink)
	for _, tc := range cases {
		dbLinks[tc.migratedStoredURL] = &db.ExternalLink{
			ID:              tc.id,
			Url:             tc.migratedStoredURL,
			Clicks:          15,
			CardTitle:       sql.NullString{String: tc.title, Valid: true},
			CardDescription: sql.NullString{String: tc.desc, Valid: true},
		}
	}

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

	cd := common.NewCoreData(context.Background(), qs, cfg, common.WithLinkSignKey(key))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Runtime canonicalization matches migrated stored URL exactly
			runtimeCanonical := common.CanonicalizeExternalURL(tc.rawTrackedURL)
			assert.Equal(t, tc.migratedStoredURL, runtimeCanonical, "runtime canonical URL must exactly match migrated URL")

			// 2. cd.GetExternalLink finds the pre-existing migrated record
			link, err := cd.GetExternalLink(context.Background(), tc.rawTrackedURL)
			assert.NoError(t, err)
			assert.NotNil(t, link)
			assert.Equal(t, tc.id, link.ID)
			assert.Equal(t, int32(15), link.Clicks)
			assert.Equal(t, tc.title, link.CardTitle.String)

			// 3. Renderer renders using the existing metadata for http/https URLs
			if strings.HasPrefix(strings.ToLower(tc.rawTrackedURL), "http://") || strings.HasPrefix(strings.ToLower(tc.rawTrackedURL), "https://") {
				provider := common.NewGoa4WebLinkProvider(cd, context.Background())
				open, close, _ := provider.RenderLink(tc.rawTrackedURL, true, true)
				rendered := open + close
				assert.Contains(t, rendered, tc.title, "must render migrated card title")
				assert.Contains(t, rendered, tc.desc, "must render migrated card description")
			}
		})
	}
}

// Compile-time check that a4code is imported and reachable
var _ = a4code.ParseString
