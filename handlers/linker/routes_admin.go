package linker

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/category/{category}", AdminCategoryPage).Methods("GET")
	lar.HandleFunc("/categories", handlers.TaskHandler(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/categories", handlers.TaskHandler(RenameCategoryTask)).Methods("POST").MatcherFunc(RenameCategoryTask.Matcher())
	lar.HandleFunc("/categories", handlers.TaskHandler(AdminDeleteCategoryTask)).Methods("POST").MatcherFunc(AdminDeleteCategoryTask.Matcher())
	lar.HandleFunc("/categories", handlers.TaskHandler(CreateCategoryTask)).Methods("POST").MatcherFunc(CreateCategoryTask.Matcher())
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", handlers.TaskHandler(AdminAddTask)).Methods("POST").MatcherFunc(AdminAddTask.Matcher())
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminDeleteTask)).Methods("POST").MatcherFunc(AdminDeleteTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminApproveTask)).Methods("POST").MatcherFunc(AdminApproveTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminBulkApproveTask)).Methods("POST").MatcherFunc(AdminBulkApproveTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(AdminBulkDeleteTask)).Methods("POST").MatcherFunc(AdminBulkDeleteTask.Matcher())
	lar.HandleFunc("/users/roles", AdminUserRolesPage).Methods("GET")
	lar.HandleFunc("/users/roles", handlers.TaskHandler(UserAllowTask)).Methods("POST").MatcherFunc(UserAllowTask.Matcher())
	lar.HandleFunc("/users/roles", handlers.TaskHandler(UserDisallowTask)).Methods("POST").MatcherFunc(UserDisallowTask.Matcher())

	lar.HandleFunc("/link/{link}", adminLinkViewPage).Methods("GET")
	lar.HandleFunc("/link/{link}/edit", adminLinkPage).Methods("GET")
	lar.HandleFunc("/link/{link}/edit", handlers.TaskHandler(AdminEditLinkTask)).Methods("POST").MatcherFunc(AdminEditLinkTask.Matcher())

	lar.HandleFunc("/category/{category}/grants", AdminCategoryGrantsPage).Methods("GET")
	lar.HandleFunc("/category/{category}/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	lar.HandleFunc("/category/{category}/grant/delete", handlers.TaskHandler(AdminCategoryGrantDeleteTask)).Methods("POST").MatcherFunc(AdminCategoryGrantDeleteTask.Matcher())
}
