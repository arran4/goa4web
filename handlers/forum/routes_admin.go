package forum

import (
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
)

// RegisterAdminRoutes attaches forum admin endpoints to ar.
func RegisterAdminRoutes(ar *mux.Router) {
	far := ar.PathPrefix("/forum").Subrouter()
	far.HandleFunc("", AdminForumPage).Methods("GET")
	far.HandleFunc("", AdminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(remakeThreadStatsTask.Matcher())
	far.HandleFunc("", AdminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(remakeTopicStatsTask.Matcher())
	far.HandleFunc("/flagged", AdminForumFlaggedPostsPage).Methods("GET")
	far.HandleFunc("/logs", AdminForumModeratorLogsPage).Methods("GET")
	far.HandleFunc("/list", AdminForumWordListPage).Methods("GET")
	far.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	far.HandleFunc("/categories", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/category/{category}", AdminCategoryEditPage).Methods("POST").MatcherFunc(categoryChangeTask.Matcher())
	far.HandleFunc("/category", AdminCategoryCreatePage).Methods("POST").MatcherFunc(categoryCreateTask.Matcher())
	far.HandleFunc("/category/delete", AdminCategoryDeletePage).Methods("POST").MatcherFunc(deleteCategoryTask.Matcher())
	far.HandleFunc("/topics", AdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/topic", AdminTopicCreatePage).Methods("POST").MatcherFunc(topicCreateTask.Matcher())
	far.HandleFunc("/conversations", AdminThreadsPage).Methods("GET")
	far.HandleFunc("/users", AdminUsersRedirect).Methods("GET")
	far.HandleFunc("/user/{id}/levels", AdminUserLevelsRedirect).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", AdminThreadDeletePage).Methods("POST").MatcherFunc(threadDeleteTask.Matcher())
	far.HandleFunc("/topic/{topic}/edit", AdminTopicEditPage).Methods("POST").MatcherFunc(topicChangeTask.Matcher())
	far.HandleFunc("/topic/{topic}/delete", AdminTopicDeletePage).Methods("POST").MatcherFunc(topicDeleteTask.Matcher())
	far.HandleFunc("/topic/{topic}/grants", AdminTopicGrantsPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/grant", handlers.TaskHandler(topicGrantCreateTask)).Methods("POST").MatcherFunc(topicGrantCreateTask.Matcher())
	far.HandleFunc("/topic/{topic}/grant/delete", handlers.TaskHandler(topicGrantDeleteTask)).Methods("POST").MatcherFunc(topicGrantDeleteTask.Matcher())

	far.HandleFunc("/category/{category}/grants", AdminCategoryGrantsPage).Methods("GET")
	far.HandleFunc("/category/{category}/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	far.HandleFunc("/category/{category}/grant/delete", handlers.TaskHandler(categoryGrantDeleteTask)).Methods("POST").MatcherFunc(categoryGrantDeleteTask.Matcher())
}
