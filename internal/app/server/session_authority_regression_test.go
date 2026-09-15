package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/sessions"
)

func TestAuthoritativeSessionRejectsUIDMismatch(t *testing.T) {
	ref := "uid-mismatch-ref"
	hash := core.HashSessionRef(ref)
	sm := &sessionManagerStub{inserted: []sessionInsert{{sessionID: hash, userID: 20}}}
	srv, store, sessionName := newAuthorityRegressionHarness(t, sm)
	cookie := authorityRegressionCookie(t, store, sessionName, int32(10), ref, 0)

	initialInsertCount := len(sm.inserted)
	authorityRejectedRequest(t, srv, cookie)
	authorityRejectedRequest(t, srv, cookie)

	if len(sm.inserted) != initialInsertCount {
		t.Fatalf("UID/SessionRef mismatch recreated session state: inserted=%d, want %d", len(sm.inserted), initialInsertCount)
	}
}

func TestAuthoritativeSessionRevokedCookieRepeatedReplay(t *testing.T) {
	ref := "revoked-replay-ref"
	hash := core.HashSessionRef(ref)
	sm := &sessionManagerStub{
		inserted: []sessionInsert{{sessionID: hash, userID: 10}},
		deleted:  []string{hash}, // simulate an administrator/server-side revocation before replay
	}
	srv, store, sessionName := newAuthorityRegressionHarness(t, sm)
	cookie := authorityRegressionCookie(t, store, sessionName, int32(10), ref, 0)

	initialInsertCount := len(sm.inserted)
	authorityRejectedRequest(t, srv, cookie)
	authorityRejectedRequest(t, srv, cookie)

	if len(sm.inserted) != initialInsertCount {
		t.Fatalf("rejected replay resurrected session state: inserted=%d, want %d", len(sm.inserted), initialInsertCount)
	}
}

func TestAuthoritativeSessionExpiryRevokesAndRejectsReplay(t *testing.T) {
	ref := "expired-authoritative-ref"
	hash := core.HashSessionRef(ref)
	sm := &sessionManagerStub{inserted: []sessionInsert{{sessionID: hash, userID: 10}}}
	srv, store, sessionName := newAuthorityRegressionHarness(t, sm)
	cookie := authorityRegressionCookie(t, store, sessionName, int32(10), ref, time.Now().Add(-time.Minute).Unix())

	initialInsertCount := len(sm.inserted)
	authorityRejectedRequest(t, srv, cookie)

	if !authorityStringPresent(sm.deleted, hash) {
		t.Fatalf("expired authoritative session was not revoked: deleted=%v, want %q", sm.deleted, hash)
	}

	// Replay the exact originally captured pre-expiry cookie. The first recovery
	// must not recreate the server-side row, and the replay must remain rejected.
	authorityRejectedRequest(t, srv, cookie)
	if len(sm.inserted) != initialInsertCount {
		t.Fatalf("expired-session recovery resurrected session state: inserted=%d, want %d", len(sm.inserted), initialInsertCount)
	}
}

func newAuthorityRegressionHarness(t *testing.T, sm *sessionManagerStub) (*Server, *sessions.CookieStore, string) {
	t.Helper()

	store := sessions.NewCookieStore([]byte("authority-regression-test-secret"))
	sessionName := "authority-regression-session"
	oldStore, oldSessionName := core.Store, core.SessionName
	core.Store, core.SessionName = store, sessionName
	t.Cleanup(func() {
		core.Store, core.SessionName = oldStore, oldSessionName
	})

	srv := New(
		WithStore(store),
		WithQuerier(testhelpers.NewQuerierStub()),
		WithConfig(config.NewRuntimeConfig()),
		WithSessionManager(sm),
	)
	return srv, store, sessionName
}

func authorityRegressionCookie(t *testing.T, store *sessions.CookieStore, sessionName string, uid int32, ref string, expiry int64) *http.Cookie {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	session, err := store.New(req, sessionName)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	session.Values["UID"] = uid
	session.Values["SessionRef"] = ref
	session.Values["LoginTime"] = time.Now().Add(-time.Hour).Unix()
	if expiry != 0 {
		session.Values["ExpiryTime"] = expiry
	}

	rr := httptest.NewRecorder()
	if err := session.Save(req, rr); err != nil {
		t.Fatalf("save session: %v", err)
	}
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == sessionName {
			return cookie
		}
	}
	t.Fatalf("response did not set %q cookie", sessionName)
	return nil
}

func authorityRejectedRequest(t *testing.T, srv *Server, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	reached := false
	handler := srv.CoreDataMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "http://example.test/usr", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if reached {
		t.Fatal("rejected authoritative session reached protected handler")
	}
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("rejected authoritative session status=%d, want 303", rr.Code)
	}
	if got := rr.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
		t.Fatalf("rejected authoritative session Cache-Control=%q, want no-cache, no-store, must-revalidate", got)
	}
	return rr
}

func authorityStringPresent(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
