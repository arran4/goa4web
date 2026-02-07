package externallink

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/testhelpers"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestRedirectHandler(t *testing.T) {
	key := "secret"

	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Redirect with go param", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			var registeredURL string
			qs.SystemRegisterExternalLinkClickFn = func(ctx context.Context, url string) error {
				registeredURL = url
				return nil
			}

			// Mock existing link to avoid DB lookup error logging
			qs.GetExternalLinkFn = func(ctx context.Context, url string) (*db.ExternalLink, error) {
				return &db.ExternalLink{ID: 1, Url: url, CardTitle: sql.NullString{String: "Title", Valid: true}, CardDescription: sql.NullString{String: "Desc", Valid: true}}, nil
			}

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key))

			urlStr := "https://example.com"
			sig := sign.Sign("link:"+urlStr, key)
			path := "/goto?u=" + urlStr + "&sig=" + sig + "&go=1"

			req := httptest.NewRequest("GET", path, nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			RedirectHandler(rec, req)

			if rec.Code != http.StatusTemporaryRedirect {
				t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
			}
			if loc := rec.Header().Get("Location"); loc != urlStr {
				t.Errorf("expected location %s, got %s", urlStr, loc)
			}
			if registeredURL != urlStr {
				t.Errorf("expected registered url %s, got %s", urlStr, registeredURL)
			}
		})

		t.Run("Render confirmation page and fetch metadata", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			// Mock GetExternalLink - return not found initially to trigger fetch
			qs.GetExternalLinkFn = func(ctx context.Context, url string) (*db.ExternalLink, error) {
				return nil, sql.ErrNoRows
			}
			// Mock CreateExternalLink
			qs.CreateExternalLinkFn = func(ctx context.Context, url string) (sql.Result, error) {
				return db.FakeSQLResult{LastInsertIDValue: 123}, nil
			}
			// Mock UpdateExternalLinkMetadata
			updatedMetadata := false
			qs.UpdateExternalLinkMetadataFn = func(ctx context.Context, arg db.UpdateExternalLinkMetadataParams) error {
				if arg.ID == 123 && arg.CardTitle.String == "Example Domain" {
					updatedMetadata = true
				}
				return nil
			}
			// Mock GetExternalLinkByID for after creation/update logic
			qs.GetExternalLinkByIDFn = func(ctx context.Context, id int32) (*db.ExternalLink, error) {
				if id == 123 {
					return &db.ExternalLink{
						ID:              123,
						Url:             "https://external.com",
						CardTitle:       sql.NullString{String: "Example Domain", Valid: true},
						CardDescription: sql.NullString{String: "Example Description", Valid: true},
					}, nil
				}
				return nil, sql.ErrNoRows
			}

			// Mock HTTP Client for OpenGraph fetch
			mockClient := NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Body: io.NopCloser(strings.NewReader(`
						<html>
							<head>
								<meta property="og:title" content="Example Domain" />
								<meta property="og:description" content="Example Description" />
							</head>
							<body></body>
						</html>
					`)),
					Header: make(http.Header),
				}
			})

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key), common.WithHTTPClient(mockClient))

			urlStr := "https://external.com"
			sig := sign.Sign("link:"+urlStr, key)
			path := "/goto?u=" + urlStr + "&sig=" + sig

			req := httptest.NewRequest("GET", path, nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			RedirectHandler(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			if !updatedMetadata {
				t.Error("expected metadata update")
			}
			// Check if page title is set in cd (CoreData is modified in place)
			if cd.PageTitle != "External Link" {
				t.Errorf("expected page title 'External Link', got '%s'", cd.PageTitle)
			}
		})
	})

	t.Run("Unhappy Path", func(t *testing.T) {
		t.Run("Missing CoreData Key", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			cd := common.NewCoreData(context.Background(), qs, nil) // No key

			req := httptest.NewRequest("GET", "/goto", nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			RedirectHandler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("Invalid Signature", func(t *testing.T) {
			qs := testhelpers.NewQuerierStub()
			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key))

			urlStr := "https://example.com"
			sig := "invalid_signature"
			path := "/goto?u=" + urlStr + "&sig=" + sig

			req := httptest.NewRequest("GET", path, nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			RedirectHandler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})

		t.Run("DB Error on GetExternalLink", func(t *testing.T) {
			// This case is actually handled gracefully in the code (logged), but if ID lookup fails it might error out?
			// The code handles GetExternalLink error: "else if err != nil && !errors.Is(err, sql.ErrNoRows) { log.Printf(...) }"
			// So it shouldn't crash or return 500, but proceed.
			// However, if we provide ID and signature valid, but DB fails to get link, it might return 400.

			qs := testhelpers.NewQuerierStub()
			qs.GetExternalLinkByIDFn = func(ctx context.Context, id int32) (*db.ExternalLink, error) {
				return nil, errors.New("db error")
			}

			cd := common.NewCoreData(context.Background(), qs, nil, common.WithLinkSignKey(key))

			idStr := "123"
			sig := sign.Sign("link:"+idStr, key)
			path := "/goto?id=" + idStr + "&sig=" + sig

			req := httptest.NewRequest("GET", path, nil)
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
			rec := httptest.NewRecorder()

			RedirectHandler(rec, req)

			// If link is not found (or error), and rawURL is empty (since we passed ID), it enters:
			// if rawURL == "" { if link == nil { error page } }
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	})
}
