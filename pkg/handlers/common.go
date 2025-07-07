package handlers

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/templates"
)

// MainCSS serves the site's stylesheet.
func MainCSS(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "main.css", time.Time{}, bytes.NewReader(templates.GetMainCSSData()))
}

// Favicon serves the site's favicon image.
func Favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	http.ServeContent(w, r, "favicon.svg", time.Time{}, bytes.NewReader(templates.GetFaviconData()))
}

// RedirectPermanent returns a handler that redirects to the provided path using StatusPermanentRedirect.
func RedirectPermanent(to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusPermanentRedirect)
	}
}

// RedirectPermanentPrefix redirects any path starting with the given prefix to the same path under a new prefix while preserving query parameters.
func RedirectPermanentPrefix(from, to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, from)
		if !strings.HasPrefix(rest, "/") && rest != "" {
			rest = "/" + rest
		}
		target := to + rest
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
	}
}
