package forum

import (
	. "github.com/arran4/gorillamuxlogic"
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
	far.HandleFunc("/topics", AdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/conversations", AdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", AdminThreadDeletePage).Methods("POST").MatcherFunc(ThreadDeleteTask.Match)
	far.HandleFunc("/topic/{topic}/edit", AdminTopicEditPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/edit", AdminTopicEditPage).Methods("POST").MatcherFunc(TopicChangeTask.Match)
	far.HandleFunc("/topic/{topic}/delete", AdminTopicDeletePage).Methods("POST").MatcherFunc(TopicDeleteTask.Match)
	far.HandleFunc("/topic/new", TopicCreatePage).Methods("GET")
	far.HandleFunc("/topic", TopicCreatePage).Methods("POST").MatcherFunc(TopicCreateTask.Match)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(UpdateTopicRestrictionTask.Match)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(SetTopicRestrictionTask.Match)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(DeleteTopicRestrictionTask.Match)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelCopyPage).Methods("POST").MatcherFunc(CopyTopicRestrictionTask.Match)
	far.HandleFunc("/users", AdminUserPage).Methods("GET")
	far.HandleFunc("/user/{user}/levels", AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(SetUserLevelTask.Match)
	far.HandleFunc("/user/{user}/levels", AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(UpdateUserLevelTask.Match)
	far.HandleFunc("/user/{user}/levels", AdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(DeleteUserLevelTask.Match)
	far.HandleFunc("/user/{user}/levels", AdminUserLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(DeleteUserLevelTask.Match)
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(UpdateUserLevelTask.Match)
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(SetUserLevelTask.Match)
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(UpdateTopicRestrictionTask.Match)
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(DeleteTopicRestrictionTask.Match)
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(SetTopicRestrictionTask.Match)
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelCopyPage).Methods("POST").MatcherFunc(CopyTopicRestrictionTask.Match)
}
