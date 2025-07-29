package writings

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/users/roles", AdminUserRolesPage).Methods("GET")
	war.HandleFunc("/users/roles", handlers.TaskHandler(userAllowTask)).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/roles", handlers.TaskHandler(userDisallowTask)).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", handlers.TaskHandler(writingCategoryChangeTask)).Methods("POST").MatcherFunc(writingCategoryChangeTask.Matcher())
	war.HandleFunc("/categories", handlers.TaskHandler(writingCategoryCreateTask)).Methods("POST").MatcherFunc(writingCategoryCreateTask.Matcher())
	war.HandleFunc("/category/{category}", AdminCategoryPage).Methods("GET")
	war.HandleFunc("/category/{category}/permissions", AdminCategoryGrantsPage).Methods("GET")
	war.HandleFunc("/category/{category}/permission", handlers.TaskHandler(writingCategoryGrantCreateTask)).Methods("POST").MatcherFunc(writingCategoryGrantCreateTask.Matcher())
	war.HandleFunc("/category/{category}/permission/delete", handlers.TaskHandler(writingCategoryGrantDeleteTask)).Methods("POST").MatcherFunc(writingCategoryGrantDeleteTask.Matcher())
	war.HandleFunc("/category/{category}/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	war.HandleFunc("/category/{category}/grant/delete", handlers.TaskHandler(categoryGrantDeleteTask)).Methods("POST").MatcherFunc(categoryGrantDeleteTask.Matcher())
}
