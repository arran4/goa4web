package externallink

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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

type myQuerier struct {
	db.Querier
}

func (q *myQuerier) CreateExternalLink(ctx context.Context, url string) (sql.Result, error) {
	return &mockResult{lastInsertId: 1}, nil
}

func (q *myQuerier) UpdateExternalLinkMetadata(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
	return nil
}

func (q *myQuerier) UpdateExternalLinkImageCache(ctx context.Context, arg db.UpdateExternalLinkImageCacheParams) error {
	return nil
}

func TestReloadExternalLinkTask_Action(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	key := "testkey"
	link := "https://example.com/some/link"
	sig := sign.Sign("link:"+link, key)

	client := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test Title"/></head><body></body></html>`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	querier := &myQuerier{}

	cd := common.NewCoreData(context.Background(), querier, cfg, func(cd *common.CoreData) {
		cd.LinkSignKey = key
	}, common.WithHTTPClient(client))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	q := req.URL.Query()
	q.Set("u", link)
	q.Set("sig", sig)
	req.URL.RawQuery = q.Encode()

	// Populate form for FormValue calls in the task
	req.ParseForm()
	req.Form = q

	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	res := reloadExternalLinkTask.Action(rec, req)

	// New behavior: returns handlers.RedirectHandler
	if _, ok := res.(handlers.RedirectHandler); !ok {
		t.Fatalf("Expected handlers.RedirectHandler, got %T", res)
	}

	val := res.(handlers.RedirectHandler)
	expectedURL := "/goto?u=" + url.QueryEscape(link) + "&sig=" + sig
	if string(val) != expectedURL {
		t.Errorf("Expected '%s', got '%s'", expectedURL, string(val))
	}
}
