package privateforum

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
