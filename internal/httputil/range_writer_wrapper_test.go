package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRangeSupport(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World! This is a test for Range requests."))
	}

	wrapped := WithRangeSupport(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Range", "bytes=0-4")
	w := httptest.NewRecorder()
	wrapped(w, req)

	if w.Code != http.StatusPartialContent {
		t.Errorf("expected 206 Partial Content, got %d", w.Code)
	}
	if w.Body.String() != "Hello" {
		t.Errorf("expected 'Hello', got %s", w.Body.String())
	}
}
