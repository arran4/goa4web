package user

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers/common"
	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterAdminRoutes attaches user admin endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	nav.RegisterAdminControlCenter("User Permissions", "/admin/users/permissions", SectionWeight-10)
	nav.RegisterAdminControlCenter("Users", "/admin/users", SectionWeight)
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users/export", adminUsersExportPage).Methods("GET")
	ar.HandleFunc("/sessions", adminSessionsPage).Methods("GET")
	ar.HandleFunc("/sessions/delete", adminSessionsDeletePage).Methods("POST")
	ar.HandleFunc("/login/attempts", adminLoginAttemptsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(common.TaskMatcher("User Allow"))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(common.TaskMatcher("User Disallow"))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsUpdatePage).Methods("POST").MatcherFunc(common.TaskMatcher("Update"))
}
