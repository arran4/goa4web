package news

import (
	"bytes"
	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPreviewRoute(t *testing.T) {
	r := mux.NewRouter()
	navReg := navpkg.NewRegistry()
	cfg := &config.RuntimeConfig{}

	RegisterRoutes(r, cfg, navReg)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test /news/preview
	url := ts.URL + "/news/preview"
	req, err := http.NewRequest("POST", url, strings.NewReader("[b Bold]"))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	bodyBuf := new(bytes.Buffer)
	bodyBuf.ReadFrom(resp.Body)
	body := bodyBuf.String()

	if !strings.Contains(body, "<strong>Bold</strong>") {
		t.Errorf("Expected '<strong>Bold</strong>', got %q", body)
	}
}

func TestPreviewHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/preview", strings.NewReader("[b Bold]"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PreviewPage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "<strong>Bold</strong>"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
