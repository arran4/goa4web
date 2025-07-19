package linker

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", UpdateCategoryTask.Action).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/categories", RenameCategoryTask.Action).Methods("POST").MatcherFunc(RenameCategoryTask.Matcher())
	lar.HandleFunc("/categories", DeleteCategoryTask.Action).Methods("POST").MatcherFunc(DeleteCategoryTask.Matcher())
	lar.HandleFunc("/categories", CreateCategoryTask.Action).Methods("POST").MatcherFunc(CreateCategoryTask.Matcher())
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", AddTask.Action).Methods("POST").MatcherFunc(AddTask.Matcher())
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", DeleteTask.Action).Methods("POST").MatcherFunc(DeleteTask.Matcher())
	lar.HandleFunc("/queue", ApproveTask.Action).Methods("POST").MatcherFunc(ApproveTask.Matcher())
	lar.HandleFunc("/queue", UpdateCategoryTask.Action).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/queue", BulkApproveTask.Action).Methods("POST").MatcherFunc(BulkApproveTask.Matcher())
	lar.HandleFunc("/queue", BulkDeleteTask.Action).Methods("POST").MatcherFunc(BulkDeleteTask.Matcher())
	lar.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", UserAllowTask.Action).Methods("POST").MatcherFunc(UserAllowTask.Matcher())
	lar.HandleFunc("/users/levels", UserDisallowTask.Action).Methods("POST").MatcherFunc(UserDisallowTask.Matcher())
}
