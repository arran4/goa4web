package privateforum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestDisablePrivateForumCaching(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		nextCalled := false
		handler := DisablePrivateForumCaching(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		}))

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/private", nil))

		if !nextCalled {
			t.Fatal("downstream handler was not called")
		}
		if got := recorder.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
			t.Errorf("Cache-Control = %q; want no-cache, no-store, must-revalidate", got)
		}
		if got := recorder.Header().Get("Cloudflare-CDN-Cache-Control"); got != "no-store" {
			t.Errorf("Cloudflare-CDN-Cache-Control = %q; want no-store", got)
		}
		if got := recorder.Header().Get("Pragma"); got != "no-cache" {
			t.Errorf("Pragma = %q; want no-cache", got)
		}
		if got := recorder.Header().Get("Expires"); got != "0" {
			t.Errorf("Expires = %q; want 0", got)
		}
	})
}

func TestRequirePrivateTopicAccess(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		if arg.Idforumtopic == 1 {
			return &db.GetForumTopicByIdForUserRow{Idforumtopic: 1}, nil
		}
		return nil, sql.ErrNoRows
	}

	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
	cd.UserID = 1

	handler := RequirePrivateTopicAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Access Granted")) // nolint:errcheck
	}))

	// Test access to Topic 1 (Granted)
	req1 := httptest.NewRequest("GET", "/topic/1", nil)
	req1 = mux.SetURLVars(req1, map[string]string{"topic": "1"})
	req1 = req1.WithContext(context.WithValue(req1.Context(), consts.KeyCoreData, cd))
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for Topic 1, got %d", rec1.Code)
	}

	// Test access to Topic 2 (Denied - NotFoundOrLogin)
	req2 := httptest.NewRequest("GET", "/topic/2", nil)
	req2 = mux.SetURLVars(req2, map[string]string{"topic": "2"})
	req2 = req2.WithContext(context.WithValue(req2.Context(), consts.KeyCoreData, cd))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found for Topic 2, got %d", rec2.Code)
	}
}
