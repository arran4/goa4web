package user

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
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
	ar.HandleFunc("/users/permissions", handlers.TaskHandler(permissionUserAllowTask)).Methods("POST").MatcherFunc(permissionUserAllowTask.Matcher())
	ar.HandleFunc("/users/permissions", handlers.TaskHandler(permissionUserDisallowTask)).Methods("POST").MatcherFunc(permissionUserDisallowTask.Matcher())
	ar.HandleFunc("/users/permissions", handlers.TaskHandler(permissionUpdateTask)).Methods("POST").MatcherFunc(permissionUpdateTask.Matcher())
}
