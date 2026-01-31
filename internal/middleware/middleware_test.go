package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/sessions"
)

func TestRedirectToLogin(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	sess := testhelpers.Must(store.New(req, "sess"))
	rr := httptest.NewRecorder()
	code := RedirectToLogin(rr, req, sess)
	if code != http.StatusSeeOther {
		t.Fatalf("code=%d", code)
	}
	if rr.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestRedirectToLoginIncludesBackAndQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/page?x=1", nil)
	rr := httptest.NewRecorder()
	RedirectToLogin(rr, req, nil)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	q := u.Query()
	if got := q.Get("back"); got != "/page?x=1" {
		t.Fatalf("back=%q", got)
	}
	if q.Has("method") {
		t.Fatalf("unexpected method param: %s", q.Get("method"))
	}
	if q.Has("data") {
		t.Fatalf("unexpected data param: %s", q.Get("data"))
	}
}

func TestRedirectToLoginPreservesPostData(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	form := url.Values{"a": {"1"}, "b": {"2"}}
	req := httptest.NewRequest(http.MethodPost, "/submit?foo=1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	sess := testhelpers.Must(store.New(req, "sess"))
	rr := httptest.NewRecorder()
	RedirectToLogin(rr, req, sess)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	q := u.Query()
	if got := q.Get("back"); got != "/submit?foo=1" {
		t.Fatalf("back=%q", got)
	}
	if got := q.Get("method"); got != http.MethodPost {
		t.Fatalf("method=%q", got)
	}
	if got := q.Get("data"); got != form.Encode() {
		t.Fatalf("data=%q", got)
	}
}
