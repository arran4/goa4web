package auth

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func roleAllowed(r *http.Request, roles ...string) bool {
	cd, ok := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	if ok && cd != nil {
		for _, lvl := range roles {
			if cd.HasRole(lvl) {
				return true
			}
		}
		return false
	}

	user, uok := r.Context().Value(hcommon.KeyUser).(*db.User)
	queries, qok := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	if !uok || !qok {
		return false
	}
	section := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
	perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(r.Context(), db.GetPermissionsByUserIdAndSectionAndSectionAllParams{
		UsersIdusers: user.Idusers,
		Section:      sql.NullString{String: section, Valid: true},
	})
	if err != nil || !perm.Level.Valid {
		return false
	}
	cd = &hcommon.CoreData{SecurityLevel: perm.Level.String}
	for _, lvl := range roles {
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}

// RequiredAccess ensures the requestor has one of the provided access levels.
func RequiredAccess(accessLevels ...string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return roleAllowed(request, accessLevels...)
	}
}

// RequiresAnAccount checks that the requester has a valid user session.
func RequiresAnAccount() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		session, err := core.GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)
		return uid != 0
	}
}
