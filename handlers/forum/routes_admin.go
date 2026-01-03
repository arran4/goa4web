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
	far.HandleFunc("/categories/category/{category}", AdminCategoryPage).Methods("GET")
	far.HandleFunc("/categories/category/{category}/edit", AdminCategoryEditPage).Methods("GET")
	far.HandleFunc("/categories/category/{category}", AdminCategoryEditSubmit).Methods("POST").MatcherFunc(categoryChangeTask.Matcher())
	far.HandleFunc("/categories/category/{category}/edit", AdminCategoryEditSubmit).Methods("POST").MatcherFunc(categoryChangeTask.Matcher())
	far.HandleFunc("/categories/category/{category}/delete", AdminCategoryDeletePage).Methods("POST").MatcherFunc(deleteCategoryTask.Matcher())
	far.HandleFunc("/categories/create", AdminCategoryCreatePage).Methods("GET")
	far.HandleFunc("/categories/create", AdminCategoryCreateSubmit).Methods("POST").MatcherFunc(categoryCreateTask.Matcher())
	far.HandleFunc("/topics", AdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", handlers.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/topic", AdminTopicCreatePage).Methods("POST").MatcherFunc(topicCreateTask.Matcher())
	far.HandleFunc("/threads", AdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}", AdminThreadPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", AdminThreadDeleteConfirmPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", AdminThreadDeletePage).Methods("POST").MatcherFunc(threadDeleteTask.Matcher())
	far.HandleFunc("/topics/topic/{topic}", AdminTopicPage).Methods("GET")
	far.HandleFunc("/topics/topic/{topic}/edit", AdminTopicEditFormPage).Methods("GET")
	far.HandleFunc("/topics/topic/{topic}/edit", AdminTopicEditPage).Methods("POST").MatcherFunc(topicChangeTask.Matcher())
	far.HandleFunc("/topics/topic/{topic}/delete", AdminTopicDeleteConfirmPage).Methods("GET")
	far.HandleFunc("/topics/topic/{topic}/delete", AdminTopicDeletePage).Methods("POST").MatcherFunc(topicDeleteTask.Matcher())
	far.HandleFunc("/topics/topic/{topic}/grants", AdminTopicGrantsPage).Methods("GET")
	far.HandleFunc("/topics/topic/{topic}/grant", handlers.TaskHandler(topicGrantCreateTask)).Methods("POST").MatcherFunc(topicGrantCreateTask.Matcher())
	far.HandleFunc("/topics/topic/{topic}/grant/update", handlers.TaskHandler(topicGrantUpdateTask)).Methods("POST").MatcherFunc(topicGrantUpdateTask.Matcher())
	far.HandleFunc("/topics/topic/{topic}/grant/delete", handlers.TaskHandler(topicGrantDeleteTask)).Methods("POST").MatcherFunc(topicGrantDeleteTask.Matcher())
	far.HandleFunc("/topic/{topic}", AdminTopicPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/edit", AdminTopicEditFormPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/edit", AdminTopicEditPage).Methods("POST").MatcherFunc(topicChangeTask.Matcher())
	far.HandleFunc("/topic/{topic}/delete", AdminTopicDeleteConfirmPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/delete", AdminTopicDeletePage).Methods("POST").MatcherFunc(topicDeleteTask.Matcher())
	far.HandleFunc("/topic/{topic}/grants", AdminTopicGrantsPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/grant", handlers.TaskHandler(topicGrantCreateTask)).Methods("POST").MatcherFunc(topicGrantCreateTask.Matcher())
	far.HandleFunc("/topic/{topic}/grant/update", handlers.TaskHandler(topicGrantUpdateTask)).Methods("POST").MatcherFunc(topicGrantUpdateTask.Matcher())
	far.HandleFunc("/topic/{topic}/grant/delete", handlers.TaskHandler(topicGrantDeleteTask)).Methods("POST").MatcherFunc(topicGrantDeleteTask.Matcher())

	far.HandleFunc("/categories/category/{category}/grants", AdminCategoryGrantsPage).Methods("GET")
	far.HandleFunc("/categories/category/{category}/grant", handlers.TaskHandler(categoryGrantCreateTask)).Methods("POST").MatcherFunc(categoryGrantCreateTask.Matcher())
	far.HandleFunc("/categories/category/{category}/grant/delete", handlers.TaskHandler(categoryGrantDeleteTask)).Methods("POST").MatcherFunc(categoryGrantDeleteTask.Matcher())
}
