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
	far.HandleFunc("", AdminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskRemakeStatisticInformationOnForumthread))
	far.HandleFunc("", AdminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskRemakeStatisticInformationOnForumtopic))
	far.HandleFunc("/flagged", AdminForumFlaggedPostsPage).Methods("GET")
	far.HandleFunc("/logs", AdminForumModeratorLogsPage).Methods("GET")
	far.HandleFunc("/list", AdminForumWordListPage).Methods("GET")
	far.HandleFunc("/categories", AdminCategoriesPage).Methods("GET")
	far.HandleFunc("/categories", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/category/{category}", AdminCategoryEditPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskForumCategoryChange))
	far.HandleFunc("/category", AdminCategoryCreatePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskForumCategoryCreate))
	far.HandleFunc("/category/delete", AdminCategoryDeletePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskDeleteCategory))
	far.HandleFunc("/topics", AdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", hcommon.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/conversations", AdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", AdminThreadDeletePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskForumThreadDelete))
	far.HandleFunc("/topic/{topic}/edit", AdminTopicEditPage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskForumTopicChange))
	far.HandleFunc("/topic/{topic}/delete", AdminTopicDeletePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskForumTopicDelete))
	far.HandleFunc("/topic", TopicCreatePage).Methods("POST").MatcherFunc(hcommon.TaskMatcher(hcommon.TaskForumTopicCreate))
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(UpdateTopicRestrictionTask.Matcher)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(SetTopicRestrictionTask.Matcher)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(DeleteTopicRestrictionTask.Matcher)
	far.HandleFunc("/topic/{topic}/levels", AdminTopicRestrictionLevelCopyPage).Methods("POST").MatcherFunc(CopyTopicRestrictionTask.Matcher)
	far.HandleFunc("/users", AdminUserPage).Methods("GET")
	far.HandleFunc("/user/{user}/levels", AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(SetUserLevelTask.Matcher)
	far.HandleFunc("/user/{user}/levels", AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(UpdateUserLevelTask.Matcher)
	far.HandleFunc("/user/{user}/levels", AdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(DeleteUserLevelTask.Matcher)
	far.HandleFunc("/user/{user}/levels", AdminUserLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(DeleteUserLevelTask.Matcher)
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(UpdateUserLevelTask.Matcher)
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(SetUserLevelTask.Matcher)
	far.HandleFunc("/restrictions/users", AdminUsersRestrictionsPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(UpdateTopicRestrictionTask.Matcher)
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(DeleteTopicRestrictionTask.Matcher)
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(SetTopicRestrictionTask.Matcher)
	far.HandleFunc("/restrictions/topics", AdminTopicsRestrictionLevelCopyPage).Methods("POST").MatcherFunc(CopyTopicRestrictionTask.Matcher)
}
