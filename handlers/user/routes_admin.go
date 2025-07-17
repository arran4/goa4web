package user

import (
	"github.com/gorilla/mux"

	nav "github.com/arran4/goa4web/internal/navigation"
)

// RegisterAdminRoutes attaches user admin endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router) {
	nav.RegisterAdminControlCenter("User Permissions", "/admin/users/permissions", SectionWeight-10)
	nav.RegisterAdminControlCenter("Pending Users", "/admin/users/pending", SectionWeight-5)
	nav.RegisterAdminControlCenter("Users", "/admin/users", SectionWeight)
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users/pending", adminPendingUsersPage).Methods("GET")
	ar.HandleFunc("/users/pending/approve", adminPendingUsersApprove).Methods("POST")
	ar.HandleFunc("/users/pending/reject", adminPendingUsersReject).Methods("POST")
	ar.HandleFunc("/users/export", adminUsersExportPage).Methods("GET")
	ar.HandleFunc("/sessions", adminSessionsPage).Methods("GET")
	ar.HandleFunc("/sessions/delete", adminSessionsDeletePage).Methods("POST")
	ar.HandleFunc("/login/attempts", adminLoginAttemptsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", permissionUserAllowTask.Action).Methods("POST").MatcherFunc(PermissionUserAllowEvent.Match)
	ar.HandleFunc("/users/permissions", permissionUserDisallowTask.Action).Methods("POST").MatcherFunc(PermissionUserDisallowEvent.Match)
	ar.HandleFunc("/users/permissions", permissionUpdateTask.Action).Methods("POST").MatcherFunc(PermissionUpdateEvent.Match)
}
