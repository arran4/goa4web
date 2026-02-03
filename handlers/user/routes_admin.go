package user

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// RegisterAdminRoutes attaches user admin endpoints to the router.
func RegisterAdminRoutes(ar *mux.Router, navReg *navpkg.Registry) {
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Users"), "Pending Users", "/admin/users/pending", SectionWeight-5)
	navReg.RegisterAdminControlCenter(navpkg.AdminCCCategory("Users"), "Users", "/admin/users", SectionWeight)
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users/pending", adminPendingUsersPage).Methods("GET")
	ar.HandleFunc("/users/pending/approve", adminPendingUsersApprove).Methods("POST")
	ar.HandleFunc("/users/pending/reject", adminPendingUsersReject).Methods("POST")
	ar.HandleFunc("/users/export", adminUsersExportPage).Methods("GET")
	ar.HandleFunc("/sessions", adminSessionsPage).Methods("GET")
	ar.HandleFunc("/sessions/delete", adminSessionsDeletePage).Methods("POST")
	ar.HandleFunc("/login/attempts", adminLoginAttemptsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/permissions", adminUserPermissionsPage).Methods("GET")
	ar.HandleFunc("/user/{user}/permissions", handlers.TaskHandler(permissionUserAllowTask)).Methods("POST").MatcherFunc(permissionUserAllowTask.Matcher())
	ar.HandleFunc("/user/{user}/permissions", handlers.TaskHandler(permissionUserDisallowTask)).Methods("POST").MatcherFunc(permissionUserDisallowTask.Matcher())
	ar.HandleFunc("/user/{user}/permissions", handlers.TaskHandler(permissionUpdateTask)).Methods("POST").MatcherFunc(permissionUpdateTask.Matcher())
	ar.HandleFunc("/user/{user}/disable", adminUserDisableConfirmPage).Methods("GET")
	ar.HandleFunc("/user/{user}/disable", adminUserDisablePage).Methods("POST")
	ar.HandleFunc("/user/{user}/edit", adminUserEditFormPage).Methods("GET")
	ar.HandleFunc("/user/{user}/edit", adminUserEditSavePage).Methods("POST")
}
