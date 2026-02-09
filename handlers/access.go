package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// DisableCaching sets headers to prevent caching of the response.
func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
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
	isPublic := false
	for _, role := range roles {
		if role == "anyone" {
			isPublic = true
			break
		}
	}
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

// EnforceGrant ensures that the user has the required grant for the item identified by the URL parameter.
func EnforceGrant(h http.HandlerFunc, section, item, action, param string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, ok := vars[param]
		if !ok {
			RenderErrorPage(w, r, ErrNotFound)
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil || id == 0 {
			RenderErrorPage(w, r, ErrNotFound)
			return
		}
		cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
		if !cd.HasGrant(section, item, action, int32(id)) {
			RenderErrorPage(w, r, ErrForbidden)
			return
		}
		h(w, r)
	}
}
