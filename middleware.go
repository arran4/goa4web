package goa4web

import (
	"database/sql"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

// routerWrapper wraps a router with additional middleware.
type routerWrapper interface {
	Wrap(http.Handler) http.Handler
}

// routerWrapperFunc allows ordinary functions to satisfy routerWrapper.
type routerWrapperFunc func(http.Handler) http.Handler

func (f routerWrapperFunc) Wrap(h http.Handler) http.Handler { return f(h) }

// newMiddlewareChain returns a routerWrapper that wraps a handler with the provided
// middleware functions in the order supplied.
func newMiddlewareChain(mw ...func(http.Handler) http.Handler) routerWrapper {
	return routerWrapperFunc(func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	})
}

// AdminCheckerMiddleware ensures the requester has administrator rights.
func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return RoleCheckerMiddleware("administrator")(next)
}

// RequestLoggerMiddleware logs incoming requests and the associated user ID.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid int32
		if u, ok := r.Context().Value(common.KeyUser).(*User); ok && u != nil {
			uid = u.Idusers
		}
		log.Printf("%s %s uid=%d", r.Method, r.URL.Path, uid)
		next.ServeHTTP(w, r)
	})
}

// roleAllowed checks if the current request has one of the provided roles.
func roleAllowed(r *http.Request, roles ...string) bool {
	cd, ok := r.Context().Value(common.KeyCoreData).(*CoreData)
	if ok && cd != nil {
		for _, lvl := range roles {
			if cd.HasRole(lvl) {
				return true
			}
		}
		return false
	}

	user, uok := r.Context().Value(common.KeyUser).(*User)
	queries, qok := r.Context().Value(common.KeyQueries).(*Queries)
	if !uok || !qok {
		return false
	}
	section := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
	perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(r.Context(), GetPermissionsByUserIdAndSectionAndSectionAllParams{
		UsersIdusers: user.Idusers,
		Section:      sql.NullString{String: section, Valid: true},
	})
	if err != nil || !perm.Level.Valid {
		return false
	}
	cd = &CoreData{SecurityLevel: perm.Level.String}
	for _, lvl := range roles {
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !roleAllowed(r, roles...) {
				err := templates.GetCompiledTemplates(corecommon.NewFuncs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", r.Context().Value(common.KeyCoreData).(*CoreData))
				if err != nil {
					log.Printf("Template Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
