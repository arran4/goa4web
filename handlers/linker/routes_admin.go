package linker

import (
	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterAdminRoutes attaches linker admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", AdminCategoriesUpdatePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUpdate))
	lar.HandleFunc("/categories", AdminCategoriesRenamePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskRenameCategory))
	lar.HandleFunc("/categories", AdminCategoriesDeletePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDeleteCategory))
	lar.HandleFunc("/categories", AdminCategoriesCreatePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskCreateCategory))
	lar.HandleFunc("/add", AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", AdminAddActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskAdd))
	lar.HandleFunc("/queue", AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", AdminQueueDeleteActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDelete))
	lar.HandleFunc("/queue", AdminQueueApproveActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskApprove))
	lar.HandleFunc("/queue", AdminQueueUpdateActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUpdate))
	lar.HandleFunc("/queue", AdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskBulkApprove))
	lar.HandleFunc("/queue", AdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskBulkDelete))
	lar.HandleFunc("/users/levels", AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserAllow))
	lar.HandleFunc("/users/levels", AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskUserDisallow))
}
