package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectPermanentPrefix(t *testing.T) {
	h := RedirectPermanentPrefix("/writing", "/writings")
	r := httptest.NewRequest("GET", "http://example.com/writing/foo?bar=baz", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusPermanentRedirect {
		t.Fatalf("want %d got %d", http.StatusPermanentRedirect, w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/writings/foo?bar=baz" {
		t.Errorf("unexpected location %s", loc)
	}
}

func TestRedirectPermanentPrefixNoMatch(t *testing.T) {
	h := RedirectPermanentPrefix("/writing", "/writings")
	r := httptest.NewRequest("GET", "http://example.com/writings", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", w.Code)
	}
}
