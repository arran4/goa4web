package linker

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", UpdateCategoryTask.Action).Methods("POST").MatcherFunc(UpdateCategoryTask.Match)
	lar.HandleFunc("/categories", RenameCategoryTask.Action).Methods("POST").MatcherFunc(RenameCategoryTask.Match)
	lar.HandleFunc("/categories", DeleteCategoryTask.Action).Methods("POST").MatcherFunc(DeleteCategoryTask.Match)
	lar.HandleFunc("/categories", CreateCategoryTask.Action).Methods("POST").MatcherFunc(CreateCategoryTask.Match)
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", AddTask.Action).Methods("POST").MatcherFunc(AddTask.Match)
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", DeleteTask.Action).Methods("POST").MatcherFunc(DeleteTask.Match)
	lar.HandleFunc("/queue", ApproveTask.Action).Methods("POST").MatcherFunc(ApproveTask.Match)
	lar.HandleFunc("/queue", UpdateCategoryTask.Action).Methods("POST").MatcherFunc(UpdateCategoryTask.Match)
	lar.HandleFunc("/queue", BulkApproveTask.Action).Methods("POST").MatcherFunc(BulkApproveTask.Match)
	lar.HandleFunc("/queue", BulkDeleteTask.Action).Methods("POST").MatcherFunc(BulkDeleteTask.Match)
	lar.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", UserAllowTask.Action).Methods("POST").MatcherFunc(UserAllowTask.Match)
	lar.HandleFunc("/users/levels", UserDisallowTask.Action).Methods("POST").MatcherFunc(UserDisallowTask.Match)
}
