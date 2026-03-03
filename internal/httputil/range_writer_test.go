package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRangeResponseWriter(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		rw := NewRangeResponseWriter(w, r)

		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("Hello, World! This is a test for Range requests."))
		rw.Serve()
	}

	// Test full request
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w1.Code)
	}
	if w1.Body.String() != "Hello, World! This is a test for Range requests." {
		t.Errorf("expected full body, got %s", w1.Body.String())
	}

	// Test Range request
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Range", "bytes=0-4")
	w2 := httptest.NewRecorder()
	handler(w2, req2)

	if w2.Code != http.StatusPartialContent {
		t.Errorf("expected 206 Partial Content, got %d", w2.Code)
	}
	if w2.Body.String() != "Hello" {
		t.Errorf("expected 'Hello', got %s", w2.Body.String())
	}
}
