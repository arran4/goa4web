package linker

import (
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", AdminCategoriesUpdatePage).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher)
	lar.HandleFunc("/categories", AdminCategoriesRenamePage).Methods("POST").MatcherFunc(RenameCategoryTask.Matcher)
	lar.HandleFunc("/categories", AdminCategoriesDeletePage).Methods("POST").MatcherFunc(DeleteCategoryTask.Matcher)
	lar.HandleFunc("/categories", AdminCategoriesCreatePage).Methods("POST").MatcherFunc(CreateCategoryTask.Matcher)
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", AdminAddActionPage).Methods("POST").MatcherFunc(AddTask.Matcher)
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", AdminQueueDeleteActionPage).Methods("POST").MatcherFunc(DeleteTask.Matcher)
	lar.HandleFunc("/queue", AdminQueueApproveActionPage).Methods("POST").MatcherFunc(ApproveTask.Matcher)
	lar.HandleFunc("/queue", AdminQueueUpdateActionPage).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher)
	lar.HandleFunc("/queue", AdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(BulkApproveTask.Matcher)
	lar.HandleFunc("/queue", AdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(BulkDeleteTask.Matcher)
	lar.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(UserAllowTask.Matcher)
	lar.HandleFunc("/users/levels", AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(UserDisallowTask.Matcher)
}
