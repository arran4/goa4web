package writings

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches writings admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("", AdminPage).Methods("GET")
	war.HandleFunc("/", AdminPage).Methods("GET")
	war.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", handlers.TaskHandler(writingCategoryCreateTask)).Methods("POST").MatcherFunc(writingCategoryCreateTask.Matcher())
	war.HandleFunc("/categories/category/{category}", AdminCategoryPage).Methods("GET")
	war.HandleFunc("/categories/category/{category}/edit", AdminCategoryEditPage).Methods("GET")
	war.HandleFunc("/categories/category/{category}/edit", handlers.TaskHandler(writingCategoryChangeTask)).Methods("POST").MatcherFunc(writingCategoryChangeTask.Matcher())
	war.HandleFunc("/categories/category/{category}/grants", AdminCategoryGrantsPage).Methods("GET")
	war.HandleFunc("/categories/category/{category}/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	war.HandleFunc("/categories/category/{category}/grant/delete", handlers.TaskHandler(categoryGrantDeleteTask)).Methods("POST").MatcherFunc(categoryGrantDeleteTask.Matcher())
}
