package linker

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestEnforceLinkerCommentsAccess(t *testing.T) {
	t.Run("Allowed", enforceLinkerCommentsAccessAllowed)
	t.Run("Denied_NoLink", enforceLinkerCommentsAccessDeniedNoLink)
	t.Run("Denied_Access", enforceLinkerCommentsAccessDeniedAccess)
}

func enforceLinkerCommentsAccessDeniedAccess(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:         1,
		ThreadID:   1,
		Title:      sql.NullString{String: "t", Valid: true},
		Listed:     sql.NullTime{Time: time.Unix(0, 0), Valid: true},
	}
	// Deny everything
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 0, sql.ErrNoRows
	}

	w, req, _ := newCommentsPageRequestWithCoreData(t, q, []string{"user"}, 2)

	handler := EnforceLinkerCommentsAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected forbidden, got %d", w.Code)
	}
}

func enforceLinkerCommentsAccessAllowed(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow = &db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserRow{
		ID:       1,
		ThreadID: 1,
		Title:    sql.NullString{String: "t", Valid: true},
		Listed:   sql.NullTime{Time: time.Unix(0, 0), Valid: true},
	}
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 1, nil // Grant Allowed
	}

	w, req, _ := newCommentsPageRequestWithCoreData(t, q, []string{"user"}, 2)

	// Simulate router
	handler := EnforceLinkerCommentsAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected access allowed, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected OK body, got %s", w.Body.String())
	}
}

func enforceLinkerCommentsAccessDeniedNoLink(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserErr = sql.ErrNoRows

	w, req, _ := newCommentsPageRequestWithCoreData(t, q, []string{"user"}, 2)

	handler := EnforceLinkerCommentsAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected forbidden, got %d", w.Code)
	}
}

// Helper similar to newCommentsPageRequest but returns req directly suitable for middleware test
func newCommentsPageRequestWithCoreData(t *testing.T, queries db.Querier, roles []string, userID int32) (*httptest.ResponseRecorder, *http.Request, *common.CoreData) {
	t.Helper()

	store := sessions.NewCookieStore([]byte("t"))
	core.Store = store
	core.SessionName = "test-session"

	req := httptest.NewRequest("GET", "/linker/comments/1", nil)
	// We MUST manually set vars because we are not using router in test
	req = mux.SetURLVars(req, map[string]string{"link": "1"})

	w := httptest.NewRecorder()
	sess := testhelpers.Must(store.Get(req, core.SessionName))
	sess.Values["UID"] = userID
	sess.Save(req, w)

	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles(roles))
	cd.UserID = userID
	// Pre-populate selections if needed, but middleware calls LoadSelectionsFromRequest
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	return w, req, cd
}

func TestEnforceLinkViewAccess(t *testing.T) {
	t.Run("Allowed", enforceLinkViewAccessAllowed)
	t.Run("Denied", enforceLinkViewAccessDenied)
}

func enforceLinkViewAccessAllowed(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 1, nil // Grant Allowed
	}

	w, req, _ := newCommentsPageRequestWithCoreData(t, q, []string{"user"}, 2)

	// Simulate router
	handler := EnforceLinkViewAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected access allowed, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected OK body, got %s", w.Body.String())
	}
}

func enforceLinkViewAccessDenied(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 0, sql.ErrNoRows // Grant Denied
	}

	w, req, _ := newCommentsPageRequestWithCoreData(t, q, []string{"user"}, 2)

	handler := EnforceLinkViewAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected forbidden, got %d", w.Code)
	}
}
