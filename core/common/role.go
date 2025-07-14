package common

import (
	"net/http"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// Context keys used by Allowed when reading from the request context.
const (
	// KeyCoreData provides access to CoreData.
	KeyCoreData ContextValues = "coreData"
	// KeyQueries holds the db.Queries pointer.
	KeyQueries ContextValues = "queries"
)

// Allowed checks if the request context provides one of the given roles.
func Allowed(r *http.Request, roles ...string) bool {
	cd, ok := r.Context().Value(KeyCoreData).(*CoreData)
	if ok && cd != nil {
		for _, lvl := range roles {
			if cd.HasRole(lvl) {
				return true
			}
		}
		return false
	}

	queries, qok := r.Context().Value(KeyQueries).(*dbpkg.Queries)
	if !qok {
		return false
	}
	var uid int32
	if cd != nil {
		uid = cd.UserID
	}
	if uid == 0 {
		return false
	}
	perms, err := queries.GetPermissionsByUserID(r.Context(), uid)
	if err != nil {
		return false
	}
	var rolesList []string
	rolesList = append(rolesList, "anonymous", "user")
	for _, p := range perms {
		if p.Role != "" {
			rolesList = append(rolesList, p.Role)
		}
	}
	cd = NewCoreData(r.Context(), queries)
	cd.SetRoles(rolesList)
	for _, lvl := range roles {
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}
