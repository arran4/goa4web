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

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestReloadExternalLinkTask(t *testing.T) {
	key := "testkey"

	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Action with URL", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			qs.CreateExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}
			qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
				if arg.ID != 123 {
					t.Errorf("Expected ID 123, got %d", arg.ID)
				}
				if arg.CardTitle.String != "Test Title" {
					t.Errorf("Expected title 'Test Title', got '%s'", arg.CardTitle.String)
				}
				return nil
			}
			qs.UpdateExternalLinkImageCacheFn = func(ctx context.Context, arg db.UpdateExternalLinkImageCacheParams) error {
				return nil
			}

			link := "https://example.com/some/link"
			sig := sign.Sign("link:"+link, key)

			client := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test Title"/></head><body></body></html>`)),
					Header:     make(http.Header),
				}
			})

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key), common.WithHTTPClient(client))

			u := "/?u=" + url.QueryEscape(link) + "&sig=" + sig
			req := httptest.NewRequest(http.MethodPost, u, nil)
			req.ParseForm()

			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)

			if _, ok := res.(handlers.RedirectHandler); !ok {
				t.Fatalf("Expected handlers.RedirectHandler, got %T", res)
			}

			val := res.(handlers.RedirectHandler)
			if string(val) != u {
				t.Errorf("Expected '%s', got '%s'", u, string(val))
			}
		})

		t.Run("Action with ID", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			link := "https://example.com/some/link"
			qs.GetExternalLinkByIDFn = func(ctx context.Context, id int32) (*db.ExternalLink, error) {
				return &db.ExternalLink{ID: 123, Url: link}, nil
			}
			qs.CreateExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				// Should find existing or just update
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}
			qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
				if arg.ID != 123 {
					t.Errorf("Expected ID 123, got %d", arg.ID)
				}
				return nil
			}

			idStr := "123"
			sig := sign.Sign("link:"+idStr, key)

			client := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test Title"/></head><body></body></html>`)),
					Header:     make(http.Header),
				}
			})

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key), common.WithHTTPClient(client))

			u := "/?id=" + idStr + "&sig=" + sig
			req := httptest.NewRequest(http.MethodPost, u, nil)
			req.ParseForm()

			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)

			if _, ok := res.(handlers.RedirectHandler); !ok {
				t.Fatalf("Expected handlers.RedirectHandler, got %T", res)
			}
		})
	})

	t.Run("Unhappy Path", func(t *testing.T) {
		t.Run("Invalid Signature", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key))

			link := "https://example.com"
			u := "/?u=" + url.QueryEscape(link) + "&sig=invalid"
			req := httptest.NewRequest(http.MethodPost, u, nil)
			req.ParseForm()
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)
			if _, ok := res.(error); !ok {
				t.Errorf("Expected error, got %T", res)
			}
		})

		t.Run("Missing LinkSignKey", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			cd := common.NewCoreData(context.Background(), qs, nil) // No key

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			res := reloadExternalLinkTask.Action(rec, req)
			if err, ok := res.(error); !ok || err.Error() != "invalid link config" {
				t.Errorf("Expected 'invalid link config', got %v", res)
			}
		})
	})
}
