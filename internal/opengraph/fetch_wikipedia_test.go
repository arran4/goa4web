package opengraph

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchWikipediaReal(t *testing.T) {
	html := `<!DOCTYPE html> <html class="client-nojs vector-feature-language-in-header-enabled vector-feature-language-in-main-menu-disabled vector-feature-language-in-main-page-header-disabled vector-feature-page-tools-pinned-disabled vector-feature-toc-pinned-clientpref-1 vector-feature-main-menu-pinned-disabled vector-feature-limited-width-clientpref-1 vector-feature-limited-width-content-enabled vector-feature-custom-font-size-clientpref-1 vector-feature-appearance-pinned-clientpref-1 skin-theme-clientpref-day vector-sticky-header-enabled vector-toc-available skin-thumbsize-clientpref-standard" lang="en" dir="ltr"> <head> <meta charset="UTF-8"> <title>Gödel, Escher, Bach - Wikipedia</title> <meta name="generator" content="MediaWiki 1.47.0-wmf.4"> <meta name="referrer" content="origin"> <meta name="referrer" content="origin-when-cross-origin"> <meta name="robots" content="max-image-preview:standard"> <meta name="format-detection" content="telephone=no"> <meta property="og:image" content="https://upload.wikimedia.org/wikipedia/commons/8/8b/Godel%2C_Escher%2C_Bach_%28first_edition%29.jpg"> <meta property="og:image:width" content="780"> <meta property="og:image:height" content="1200"> <meta name="viewport" content="width=1120"> <meta property="og:title" content="Gödel, Escher, Bach - Wikipedia"> <meta property="og:type" content="website"> </head>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer server.Close()

	info, err := Fetch(server.URL, http.DefaultClient)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
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
