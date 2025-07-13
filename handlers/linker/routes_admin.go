package linker

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", AdminCategoriesUpdatePage).Methods("POST").MatcherFunc(UpdateCategoryTask.Match)
	lar.HandleFunc("/categories", AdminCategoriesRenamePage).Methods("POST").MatcherFunc(RenameCategoryTask.Match)
	lar.HandleFunc("/categories", AdminCategoriesDeletePage).Methods("POST").MatcherFunc(DeleteCategoryTask.Match)
	lar.HandleFunc("/categories", AdminCategoriesCreatePage).Methods("POST").MatcherFunc(CreateCategoryTask.Match)
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", AdminAddActionPage).Methods("POST").MatcherFunc(AddTask.Match)
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", AdminQueueDeleteActionPage).Methods("POST").MatcherFunc(DeleteTask.Match)
	lar.HandleFunc("/queue", AdminQueueApproveActionPage).Methods("POST").MatcherFunc(ApproveTask.Match)
	lar.HandleFunc("/queue", AdminQueueUpdateActionPage).Methods("POST").MatcherFunc(UpdateCategoryTask.Match)
	lar.HandleFunc("/queue", AdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(BulkApproveTask.Match)
	lar.HandleFunc("/queue", AdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(BulkDeleteTask.Match)
	lar.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(UserAllowTask.Match)
	lar.HandleFunc("/users/levels", AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(UserDisallowTask.Match)
}
