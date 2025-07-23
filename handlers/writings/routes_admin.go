package writings

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", handlers.TaskHandler(userAllowTask)).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/permissions", handlers.TaskHandler(userDisallowTask)).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/users/roles", AdminUserRolesPage).Methods("GET")
	war.HandleFunc("/users/roles", handlers.TaskHandler(userAllowTask)).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/roles", handlers.TaskHandler(userDisallowTask)).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", handlers.TaskHandler(writingCategoryChangeTask)).Methods("POST").MatcherFunc(writingCategoryChangeTask.Matcher())
	war.HandleFunc("/categories", handlers.TaskHandler(writingCategoryCreateTask)).Methods("POST").MatcherFunc(writingCategoryCreateTask.Matcher())
}
