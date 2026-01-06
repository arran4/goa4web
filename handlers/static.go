package handlers

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
)

// MainCSS serves the site's stylesheet.
func MainCSS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		http.ServeContent(w, r, "main.css", time.Time{}, bytes.NewReader(templates.GetMainCSSData(opts...)))
	}
}

// Favicon serves the site's favicon image.
func Favicon(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeContent(w, r, "favicon.svg", time.Time{}, bytes.NewReader(templates.GetFaviconData(opts...)))
	}
}

// PasteImageJS serves the JavaScript enabling clipboard image pasting.
func PasteImageJS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, r, "pasteimg.js", time.Time{}, bytes.NewReader(templates.GetPasteImageJSData(opts...)))
	}
}

// RoleGrantsEditorJS serves the JavaScript for the role grants editor.
func RoleGrantsEditorJS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, r, "role_grants_editor.js", time.Time{}, bytes.NewReader(templates.GetRoleGrantsEditorJSData(opts...)))
	}
}

// PrivateForumJS serves the JavaScript for the private forum pages.
func PrivateForumJS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, r, "private_forum.js", time.Time{}, bytes.NewReader(templates.GetPrivateForumJSData(opts...)))
	}
}

// TopicLabelsJS serves the JavaScript for topic label editing.
func TopicLabelsJS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, r, "topic_labels.js", time.Time{}, bytes.NewReader(templates.GetTopicLabelsJSData(opts...)))
	}
}

// SiteJS serves the main site JavaScript.
func SiteJS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, r, "site.js", time.Time{}, bytes.NewReader(templates.GetSiteJSData(opts...)))
	}
}

// A4CodeJS serves the A4Code parser/converter JavaScript.
func A4CodeJS(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts []any
		if cfg != nil && cfg.TemplatesDir != "" {
			opts = append(opts, cfg.TemplatesDir)
		}
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeContent(w, r, "a4code.js", time.Time{}, bytes.NewReader(templates.GetA4CodeJSData(opts...)))
	}
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
			// not an exact match or subpath - avoid redirect loop
			http.NotFound(w, r)
			return
		}
		target := to + rest
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
	}
}
