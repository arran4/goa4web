package handlers

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
)

// StaticAssetHandler returns a function that generates an http.HandlerFunc for the given asset.
func StaticAssetHandler(filename, contentType string, getter func(opts ...templates.Option) []byte) func(cfg *config.RuntimeConfig) http.HandlerFunc {
	return func(cfg *config.RuntimeConfig) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var opts []templates.Option
			if cfg != nil && cfg.TemplatesDir != "" {
				opts = append(opts, templates.WithDir(cfg.TemplatesDir))
			}
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			http.ServeContent(w, r, filename, time.Time{}, bytes.NewReader(getter(opts...)))
		}
	}
}

var (
	// MainCSS serves the site's stylesheet.
	MainCSS = StaticAssetHandler("main.css", "", templates.GetMainCSSData)

	// Favicon serves the site's favicon image.
	Favicon = StaticAssetHandler("favicon.svg", "image/svg+xml", templates.GetFaviconData)

	// PasteImageJS serves the JavaScript enabling clipboard image pasting.
	PasteImageJS = StaticAssetHandler("pasteimg.js", "application/javascript", templates.GetPasteImageJSData)

	// RoleGrantsEditorJS serves the JavaScript for the role grants editor.
	RoleGrantsEditorJS = StaticAssetHandler("role_grants_editor.js", "application/javascript", templates.GetRoleGrantsEditorJSData)

	// GrantAddJS serves the JavaScript for the admin grant add page.
	GrantAddJS = StaticAssetHandler("grant_add.js", "application/javascript", templates.GetGrantAddJSData)

	// PrivateForumJS serves the JavaScript for the private forum pages.
	PrivateForumJS = StaticAssetHandler("private_forum.js", "application/javascript", templates.GetPrivateForumJSData)

	// TopicLabelsJS serves the JavaScript for topic label editing.
	TopicLabelsJS = StaticAssetHandler("topic_labels.js", "application/javascript", templates.GetTopicLabelsJSData)

	// SiteJS serves the main site JavaScript.
	SiteJS = StaticAssetHandler("site.js", "application/javascript", templates.GetSiteJSData)

	// PasskeysJS serves the JavaScript for passkey registration and login.
	PasskeysJS = StaticAssetHandler("passkeys.js", "application/javascript", templates.GetPasskeysJSData)

	// A4CodeJS serves the A4Code parser/converter JavaScript.
	A4CodeJS = StaticAssetHandler("a4code.js", "application/javascript", templates.GetA4CodeJSData)

	// RobotsTXT serves the robots.txt file.
	RobotsTXT = StaticAssetHandler("robots.txt", "text/plain", templates.GetRobotsTXTData)
)

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
