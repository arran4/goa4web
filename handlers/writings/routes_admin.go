package writings

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", userAllowTask.Action).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/permissions", userDisallowTask.Action).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", userAllowTask.Action).Methods("POST").MatcherFunc(userAllowTask.Matcher())
	war.HandleFunc("/users/levels", userDisallowTask.Action).Methods("POST").MatcherFunc(userDisallowTask.Matcher())
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", writingCategoryChangeTask.Action).Methods("POST").MatcherFunc(writingCategoryChangeTask.Matcher())
	war.HandleFunc("/categories", writingCategoryCreateTask.Action).Methods("POST").MatcherFunc(writingCategoryCreateTask.Matcher())
}
