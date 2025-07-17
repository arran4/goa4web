package common

import (
	"net/http"

	dbpkg "github.com/arran4/goa4web/internal/db"
)

// Context keys used by Allowed when reading from the request context.

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
		if p.Name != "" {
			rolesList = append(rolesList, p.Name)
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
