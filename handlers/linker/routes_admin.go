package linker

import (
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", tasks.Action(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/categories", tasks.Action(RenameCategoryTask)).Methods("POST").MatcherFunc(RenameCategoryTask.Matcher())
	lar.HandleFunc("/categories", tasks.Action(DeleteCategoryTask)).Methods("POST").MatcherFunc(DeleteCategoryTask.Matcher())
	lar.HandleFunc("/categories", tasks.Action(CreateCategoryTask)).Methods("POST").MatcherFunc(CreateCategoryTask.Matcher())
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", tasks.Action(AddTask)).Methods("POST").MatcherFunc(AddTask.Matcher())
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", tasks.Action(DeleteTask)).Methods("POST").MatcherFunc(DeleteTask.Matcher())
	lar.HandleFunc("/queue", tasks.Action(ApproveTask)).Methods("POST").MatcherFunc(ApproveTask.Matcher())
	lar.HandleFunc("/queue", tasks.Action(UpdateCategoryTask)).Methods("POST").MatcherFunc(UpdateCategoryTask.Matcher())
	lar.HandleFunc("/queue", tasks.Action(BulkApproveTask)).Methods("POST").MatcherFunc(BulkApproveTask.Matcher())
	lar.HandleFunc("/queue", tasks.Action(BulkDeleteTask)).Methods("POST").MatcherFunc(BulkDeleteTask.Matcher())
	lar.HandleFunc("/users/roles", AdminUserRolesPage).Methods("GET")
	lar.HandleFunc("/users/roles", tasks.Action(UserAllowTask)).Methods("POST").MatcherFunc(UserAllowTask.Matcher())
	lar.HandleFunc("/users/roles", tasks.Action(UserDisallowTask)).Methods("POST").MatcherFunc(UserDisallowTask.Matcher())
}
