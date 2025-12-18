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

// PasteImageJS serves the JavaScript enabling clipboard image pasting.
func PasteImageJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeContent(w, r, "pasteimg.js", time.Time{}, bytes.NewReader(templates.GetPasteImageJSData()))
}

// RoleGrantsEditorJS serves the JavaScript for the role grants editor.
func RoleGrantsEditorJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeContent(w, r, "role_grants_editor.js", time.Time{}, bytes.NewReader(templates.GetRoleGrantsEditorJSData()))
}

// PrivateForumJS serves the JavaScript for the private forum pages.
func PrivateForumJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeContent(w, r, "private_forum.js", time.Time{}, bytes.NewReader(templates.GetPrivateForumJSData()))
}

// TopicLabelsJS serves the JavaScript for topic label editing.
func TopicLabelsJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeContent(w, r, "topic_labels.js", time.Time{}, bytes.NewReader(templates.GetTopicLabelsJSData()))
}

// SiteJS serves the main site JavaScript.
func SiteJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeContent(w, r, "site.js", time.Time{}, bytes.NewReader(templates.GetSiteJSData()))
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
