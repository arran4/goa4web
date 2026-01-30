package externallink

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
)

type mockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

type mockResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (m *mockResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}
func (m *mockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

type mockQuerier struct {
	db.Querier
}

func (q *mockQuerier) GetExternalLink(ctx context.Context, url string) (*db.ExternalLink, error) {
	return nil, sql.ErrNoRows
}

func (q *mockQuerier) GetExternalLinkByID(ctx context.Context, id int32) (*db.ExternalLink, error) {
	return &db.ExternalLink{
		ID:        id,
		Url:       "https://external.com/reload",
		CardTitle: sql.NullString{String: "Test Title", Valid: true},
	}, nil
}

func (q *mockQuerier) CreateExternalLink(ctx context.Context, url string) (sql.Result, error) {
	return &mockResult{lastInsertId: 123}, nil
}

func (q *mockQuerier) UpdateExternalLinkMetadata(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
	return nil
}

func (q *mockQuerier) UpdateExternalLinkImageCache(ctx context.Context, arg db.UpdateExternalLinkImageCacheParams) error {
	return nil
}

func TestRedirectHandlerSignedURLParam(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	key := "k"
	link := "https://example.com/foo"
	sig := sign.Sign("link:"+link, key)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/goto?u=%s&sig=%s&go=1", url.QueryEscape(link), sig), nil)
	cd := common.NewCoreData(context.Background(), nil, cfg, func(cd *common.CoreData) {
		cd.LinkSignKey = key
	})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	RedirectHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}
	if got := res.Header.Get("Location"); got != link {
		t.Fatalf("expected redirect to %s, got %s", link, got)
	}
}

func TestRedirectHandlerSignedURLParamWithQuery(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	key := "k"
	link := "https://example.com?a=1&b=2"
	sig := sign.Sign("link:"+link, key)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/goto?u=%s&sig=%s&go=1", url.QueryEscape(link), sig), nil)
	cd := common.NewCoreData(context.Background(), nil, cfg, func(cd *common.CoreData) {
		cd.LinkSignKey = key
	})
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	RedirectHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}
	if got := res.Header.Get("Location"); got != link {
		t.Fatalf("expected redirect to %s, got %s", link, got)
	}
}
