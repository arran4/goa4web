package opengraph

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
)

func TestFetchWikipedia(t *testing.T) {
	html := `<!DOCTYPE html><html><head><title>Gödel, Escher, Bach - Wikipedia</title><meta property="og:image" content="https://upload.wikimedia.org/wikipedia/commons/8/8b/Godel%2C_Escher%2C_Bach_%28first_edition%29.jpg"><meta property="og:title" content="Gödel, Escher, Bach - Wikipedia"></head>`

	info, err := Parse(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if info.Title != "Gödel, Escher, Bach - Wikipedia" {
		t.Errorf("Title = %q, want %q", info.Title, "Gödel, Escher, Bach - Wikipedia")
	}
	if info.Image != "https://upload.wikimedia.org/wikipedia/commons/8/8b/Godel%2C_Escher%2C_Bach_%28first_edition%29.jpg" {
		t.Errorf("Image = %q", info.Image)
	}
}

func TestFetchUserAgent(t *testing.T) {
	var userAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent = r.Header.Get("User-Agent")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Test</title></head></html>`))
	}))
	defer server.Close()

	_, err := Fetch(server.URL, http.DefaultClient)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if userAgent != "goa4web/1.0 (+https://github.com/arran4/goa4web)" {
		t.Errorf("User-Agent = %q, want %q", userAgent, "goa4web/1.0 (+https://github.com/arran4/goa4web)")
	}
}
