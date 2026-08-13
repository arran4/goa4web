package handlers

import (
	"net/http"
	"slices"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// DisableCaching sets headers to prevent caching of the response, including Cloudflare-specific instructions.
func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Cloudflare-CDN-Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// WithNoCache wraps a handler and ensures no-cache headers are set.
func WithNoCache(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		DisableCaching(w)
		h(w, r)
	}
}

func RenderPermissionDenied(w http.ResponseWriter, r *http.Request) {
	DisableCaching(w)
	RenderErrorPage(w, r, WrapForbidden(ErrLoginRequired))
}

// RequireRole wraps h and denies the request if the caller lacks any of
// the provided roles. err is shown on the rendered error page; if err is nil
// a generic "Forbidden" message is displayed.
func RequireRole(h http.HandlerFunc, err error, roles ...string) http.HandlerFunc {
	if err == nil {
		err = ErrForbidden
	}
	isPublic := slices.Contains(roles, "anyone")
	return func(w http.ResponseWriter, r *http.Request) {
		if !isPublic {
			DisableCaching(w)
		}
		if !common.Allowed(r, roles...) {
			w.WriteHeader(http.StatusForbidden)
			RenderErrorPage(w, r, err)
			return
		}
		h(w, r)
	}
}

// RenderNotFoundOrLogin renders the 404 page if the user is logged in,
// otherwise it renders the permission denied (login) page.
func RenderNotFoundOrLogin(w http.ResponseWriter, r *http.Request) {
	DisableCaching(w)
	cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd != nil && cd.UserID != 0 {
		RenderErrorPage(w, r, ErrNotFound)
	} else {
		RenderPermissionDenied(w, r)
	}
}

// MaximumCache sets headers to cache the response for a long time (1 year), including Cloudflare specific caches.
func MaximumCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Cloudflare-CDN-Cache-Control", "public, max-age=31536000, immutable")
}

// WithMaximumCache wraps a handler and ensures maximum caching headers are set.
func WithMaximumCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MaximumCache(w)
		h.ServeHTTP(w, r)
	})
}

// WithMaximumCacheFunc wraps a HandlerFunc and ensures maximum caching headers are set.
func WithMaximumCacheFunc(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		MaximumCache(w)
		h(w, r)
	}
}
