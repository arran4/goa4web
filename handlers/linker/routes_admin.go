package linker

import (
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", handlers.TaskHandler(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/categories", handlers.TaskHandler(RenameCategoryTask)).Methods("POST").MatcherFunc(RenameCategoryTask.Matcher())
	lar.HandleFunc("/categories", handlers.TaskHandler(DeleteCategoryTask)).Methods("POST").MatcherFunc(DeleteCategoryTask.Matcher())
	lar.HandleFunc("/categories", handlers.TaskHandler(CreateCategoryTask)).Methods("POST").MatcherFunc(CreateCategoryTask.Matcher())
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", handlers.TaskHandler(AddTask)).Methods("POST").MatcherFunc(AddTask.Matcher())
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", handlers.TaskHandler(DeleteTask)).Methods("POST").MatcherFunc(DeleteTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(ApproveTask)).Methods("POST").MatcherFunc(ApproveTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(BulkApproveTask)).Methods("POST").MatcherFunc(BulkApproveTask.Matcher())
	lar.HandleFunc("/queue", handlers.TaskHandler(BulkDeleteTask)).Methods("POST").MatcherFunc(BulkDeleteTask.Matcher())
	lar.HandleFunc("/users/roles", AdminUserRolesPage).Methods("GET")
	lar.HandleFunc("/users/roles", handlers.TaskHandler(UserAllowTask)).Methods("POST").MatcherFunc(UserAllowTask.Matcher())
	lar.HandleFunc("/users/roles", handlers.TaskHandler(UserDisallowTask)).Methods("POST").MatcherFunc(UserDisallowTask.Matcher())

	lar.HandleFunc("/category/{category}/grants", AdminCategoryGrantsPage).Methods("GET")
	lar.HandleFunc("/category/{category}/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	lar.HandleFunc("/category/{category}/grant/delete", handlers.TaskHandler(categoryGrantDeleteTask)).Methods("POST").MatcherFunc(categoryGrantDeleteTask.Matcher())
}
