package forum

import (
	"github.com/gorilla/mux"

	hcommon "github.com/arran4/goa4web/handlers/common"
)

// RegisterAdminRoutes attaches forum admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	far := ar.PathPrefix("/forum").Subrouter()
	far.HandleFunc("", AdminForumPage).Methods("GET")
	far.HandleFunc("", AdminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(RemakeThreadStatsTask.Match)
	far.HandleFunc("", AdminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(RemakeTopicStatsTask.Match)
	far.HandleFunc("/flagged", AdminForumFlaggedPostsPage).Methods("GET")
	far.HandleFunc("/logs", AdminForumModeratorLogsPage).Methods("GET")
	far.HandleFunc("/list", AdminForumWordListPage).Methods("GET")
	far.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	far.HandleFunc("/categories", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/category/{category}", AdminCategoryEditPage).Methods("POST").MatcherFunc(CategoryChangeTask.Match)
	far.HandleFunc("/category", AdminCategoryCreatePage).Methods("POST").MatcherFunc(CategoryCreateTask.Match)
	far.HandleFunc("/category/delete", AdminCategoryDeletePage).Methods("POST").MatcherFunc(DeleteCategoryTask.Match)
	far.HandleFunc("/topics", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/conversations", AdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", AdminThreadDeletePage).Methods("POST").MatcherFunc(ThreadDeleteTask.Match)
}
