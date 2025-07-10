package permissions

import (
	"database/sql"
	"net/http"
	"strings"

	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

// Allowed checks if the request context provides one of the given roles.
func Allowed(r *http.Request, roles ...string) bool {
	cd, ok := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData)
	if ok && cd != nil {
		for _, lvl := range roles {
			if cd.HasRole(lvl) {
				return true
			}
		}
		return false
	}

	queries, qok := r.Context().Value(hcommon.KeyQueries).(*dbpkg.Queries)
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
	section := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
	perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(r.Context(), dbpkg.GetPermissionsByUserIdAndSectionAndSectionAllParams{
		UsersIdusers: uid,
		Section:      sql.NullString{String: section, Valid: true},
	})
	if err != nil || !perm.Level.Valid {
		return false
	}
	cd = &corecommon.CoreData{SecurityLevel: perm.Level.String}
	for _, lvl := range roles {
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}
