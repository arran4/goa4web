package goa4web

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
	"net/http"

	"github.com/arran4/goa4web/handlers/common"
	faq "github.com/arran4/goa4web/handlers/faq"

	"github.com/arran4/goa4web/pkg/handlers"
	"github.com/arran4/goa4web/runtimeconfig"
)

func registerRoutes(r *mux.Router) {
	r.HandleFunc("/main.css", handlers.MainCSS).Methods("GET")

	registerNewsRoutes(r)
	faq.RegisterRoutes(r)
	registerBlogsRoutes(r)
	registerForumRoutes(r)
	registerLinkerRoutes(r)
	registerBookmarksRoutes(r)
	registerImagebbsRoutes(r)
	registerSearchRoutes(r)
	registerWritingsRoutes(r)
	registerInformationRoutes(r)
	registerUserRoutes(r)
	registerRegisterRoutes(r)
	registerLoginRoutes(r)
	registerAdminRoutes(r)

	// legacy redirects
	r.PathPrefix("/writing").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
	r.PathPrefix("/links").HandlerFunc(handlers.RedirectPermanentPrefix("/links", "/linker"))
}

func registerNewsRoutes(r *mux.Router) {
	// News
	r.Handle("/", AddNewsIndex(http.HandlerFunc(runTemplate("page.gohtml")))).Methods("GET")
	r.HandleFunc("/", common.TaskDoneAutoRefreshPage).Methods("POST")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(AddNewsIndex)
	nr.HandleFunc(".rss", newsRssPage).Methods("GET")
	nr.HandleFunc("", runTemplate("page.gohtml")).Methods("GET")
	nr.HandleFunc("", common.TaskDoneAutoRefreshPage).Methods("POST")
	//TODO nr.HandleFunc("/news/{id:[0-9]+}", newsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", newsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", newsPostReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskReply))
	nr.HandleFunc("/news/{post}", newsPostEditActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskEdit))
	nr.HandleFunc("/news/{post}", newsPostNewActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskNewPost))
	nr.HandleFunc("/news/{post}/announcement", newsAnnouncementActivateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskAdd))
	nr.HandleFunc("/news/{post}/announcement", newsAnnouncementDeactivateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskDelete))
	nr.HandleFunc("/news/{post}", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	nr.HandleFunc("/news/{post}", common.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", newsUserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", newsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Allow"))
	nr.HandleFunc("/users/permissions", newsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Disallow"))
}

func registerBlogsRoutes(r *mux.Router) {
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", blogsRssPage).Methods("GET")
	br.HandleFunc("/atom", blogsAtomPage).Methods("GET")
	br.HandleFunc("", blogsPage).Methods("GET")
	br.HandleFunc("/", blogsPage).Methods("GET")
	br.HandleFunc("/add", blogsBlogAddPage).Methods("GET").MatcherFunc(RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", blogsBlogAddActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskAdd))
	br.HandleFunc("/bloggers", blogsBloggersPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", blogsBloggerPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", blogsBloggerPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", blogsBlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", common.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", blogsCommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", blogsBlogReplyPostPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogsCommentEditPostPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher(TaskEditReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogsCommentEditPostCancelPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	br.HandleFunc("/blog/{blog}/edit", blogsBlogEditPage).Methods("GET").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor())))
	br.HandleFunc("/blog/{blog}/edit", blogsBlogEditActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor()))).MatcherFunc(TaskMatcher(TaskEdit))
	br.HandleFunc("/blog/{blog}/edit", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", getPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", blogsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserAllow))
	br.HandleFunc("/users/permissions", blogsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserDisallow))
	br.HandleFunc("/users/permissions", blogsUsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUsersAllow))
	br.HandleFunc("/users/permissions", blogsUsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUsersDisallow))
}

func registerForumRoutes(r *mux.Router) {
	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("/topic/{topic}.rss", forumTopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", forumTopicAtomPage).Methods("GET")
	fr.HandleFunc("", forumPage).Methods("GET")
	fr.HandleFunc("/category/{category}", forumPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}", forumTopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateThread))
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewCancelPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}", forumThreadPage).Methods("GET").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskReply))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forumTopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskEditReply)).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor()))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forumTopicThreadCommentEditActionCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskCancel))
}

func registerLinkerRoutes(r *mux.Router) {
	lr := r.PathPrefix("/linker").Subrouter()
	lr.HandleFunc("/rss", linkerRssPage).Methods("GET")
	lr.HandleFunc("/atom", linkerAtomPage).Methods("GET")
	lr.HandleFunc("", linkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}", linkerLinkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", linkerLinkerPage).Methods("GET")
	lr.HandleFunc("/categories", linkerCategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", linkerCategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linkerCommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linkerCommentsReplyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	lr.HandleFunc("/show/{link}", linkerShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", linkerShowReplyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	lr.HandleFunc("/suggest", linkerSuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", linkerSuggestActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSuggest))
}

func registerBookmarksRoutes(r *mux.Router) {
	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarksPage).Methods("GET")
	bmr.HandleFunc("/mine", bookmarksMinePage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSave))
	bmr.HandleFunc("/edit", bookmarksEditCreateActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskCreate))
	bmr.HandleFunc("/edit", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiresAnAccount())
}

func registerImagebbsRoutes(r *mux.Router) {
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(runtimeconfig.AppRuntimeConfig.ImageUploadDir))))
	ibr.HandleFunc(".rss", imagebbsRssPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", imagebbsBoardRssPage).Methods("GET")
	ibr.HandleFunc(".atom", imagebbsAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", imagebbsBoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbsBoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbsBoardPostImageActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskUploadImage))
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbsBoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbsBoardThreadReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskReply))
	ibr.HandleFunc("", imagebbsPage).Methods("GET")
	ibr.HandleFunc("/", imagebbsPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}", imagebbsPosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", imagebbsPosterPage).Methods("GET")

	// Admin endpoints for image boards
	ibr.HandleFunc("/admin", imagebbsAdminPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", imagebbsAdminBoardsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbsAdminNewBoardPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbsAdminNewBoardMakePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskNewBoard))
	ibr.HandleFunc("/admin/board", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board/{board}", imagebbsAdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskModifyBoard))
	ibr.HandleFunc("/admin/approve/{post}", imagebbsAdminApprovePostPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskApprove))
	ibr.HandleFunc("/admin/files", imagebbsAdminFilesPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
}

func registerSearchRoutes(r *mux.Router) {
	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", searchPage).Methods("GET")
	sr.HandleFunc("", searchResultForumActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchForum))
	sr.HandleFunc("", searchResultNewsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchNews))
	sr.HandleFunc("", searchResultLinkerActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchLinker))
	sr.HandleFunc("", searchResultBlogsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchBlogs))
	sr.HandleFunc("", searchResultWritingsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchWritings))
}

func registerWritingsRoutes(r *mux.Router) {
	wr := r.PathPrefix("/writings").Subrouter()
	wr.HandleFunc("/rss", writingsRssPage).Methods("GET")
	wr.HandleFunc("/atom", writingsAtomPage).Methods("GET")
	wr.HandleFunc("", writingsPage).Methods("GET")
	wr.HandleFunc("/", writingsPage).Methods("GET")
	wr.HandleFunc("/writer/{username}", writingsWriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", writingsWriterPage).Methods("GET")
	wr.HandleFunc("/user/permissions", writingsUserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", writingsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserAllow))
	wr.HandleFunc("/users/permissions", writingsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserDisallow))
	wr.HandleFunc("/article/{article}", writingsArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", writingsArticleReplyActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	wr.HandleFunc("/article/{article}/edit", writingsArticleEditPage).Methods("GET").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator")))
	wr.HandleFunc("/article/{article}/edit", writingsArticleEditActionPage).Methods("POST").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher(TaskUpdateWriting))
	wr.HandleFunc("/categories", writingsCategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", writingsCategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", writingsCategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", writingsArticleAddPage).Methods("GET").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", writingsArticleAddActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher(TaskSubmitWriting))
}

func registerInformationRoutes(r *mux.Router) {
	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", informationPage).Methods("GET")
}

func registerUserRoutes(r *mux.Router) {
	ur := r.PathPrefix("/usr").Subrouter()
	ur.HandleFunc("", userPage).Methods("GET")
	ur.HandleFunc("/logout", userLogoutPage).Methods("GET")
	ur.HandleFunc("/lang", userLangPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/lang", userLangSaveLanguagesActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveLanguages))
	ur.HandleFunc("/lang", userLangSaveLanguagePreferenceActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveLanguage))
	ur.HandleFunc("/lang", userLangSaveAllActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/email", userEmailSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailTestActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskTestMail))
	ur.HandleFunc("/paging", userPagingPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/paging", userPagingSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/page-size", userPageSizePage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/page-size", userPageSizeSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/notifications", userNotificationsPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/notifications/dismiss", userNotificationsDismissActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskDismiss))
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods("GET").MatcherFunc(RequiresAnAccount())

	// legacy redirects
	r.HandleFunc("/user/lang", handlers.RedirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", handlers.RedirectPermanent("/usr/email"))
}

func registerRegisterRoutes(r *mux.Router) {
	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", registerPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", registerActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(TaskMatcher(TaskRegister))
}

func registerLoginRoutes(r *mux.Router) {
	ulr := r.PathPrefix("/login").Subrouter()
	ulr.HandleFunc("", loginUserPassPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	ulr.HandleFunc("", loginActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(TaskMatcher(TaskLogin))
}

func registerAdminRoutes(r *mux.Router) {
	ar := r.PathPrefix("/admin").Subrouter()
	ar.Use(AdminCheckerMiddleware)
	ar.HandleFunc("", adminPage).Methods("GET")
	ar.HandleFunc("/", adminPage).Methods("GET")
	ar.HandleFunc("/categories", adminCategoriesPage).Methods("GET")
	ar.HandleFunc("/permissions/sections", adminPermissionsSectionPage).Methods("GET")
	ar.HandleFunc("/permissions/sections/view", adminPermissionsSectionViewPage).Methods("GET")
	ar.HandleFunc("/permissions/sections", adminPermissionsSectionRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameSection))
	ar.HandleFunc("/email/queue", adminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/queue", adminEmailQueueResendActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskResend))
	ar.HandleFunc("/email/queue", adminEmailQueueDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	ar.HandleFunc("/email/template", adminEmailTemplatePage).Methods("GET")
	ar.HandleFunc("/email/template", adminEmailTemplateSaveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	ar.HandleFunc("/email/template", adminEmailTemplateTestActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskTestMail))
	ar.HandleFunc("/notifications", adminNotificationsPage).Methods("GET")
	ar.HandleFunc("/notifications", adminNotificationsMarkReadActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDismiss))
	ar.HandleFunc("/notifications", adminNotificationsPurgeActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskPurge))
	ar.HandleFunc("/notifications", adminNotificationsSendActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskNotify))
	ar.HandleFunc("/announcements", adminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", adminAnnouncementsAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAdd))
	ar.HandleFunc("/announcements", adminAnnouncementsDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	ar.HandleFunc("/sessions", adminSessionsPage).Methods("GET")
	ar.HandleFunc("/sessions/delete", adminSessionsDeletePage).Methods("POST")
	ar.HandleFunc("/login/attempts", adminLoginAttemptsPage).Methods("GET")
	ar.HandleFunc("/ipbans", adminIPBanPage).Methods("GET")
	ar.HandleFunc("/ipbans", adminIPBanAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAdd))
	ar.HandleFunc("/ipbans", adminIPBanDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users/export", adminUsersExportPage).Methods("GET")
	ar.HandleFunc("/audit", adminAuditLogPage).Methods("GET")
	ar.HandleFunc("/settings", adminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/stats", adminServerStatsPage).Methods("GET")
	ar.HandleFunc("/usage", adminUsageStatsPage).Methods("GET")

	// search related
	ar.HandleFunc("/search", adminSearchPage).Methods("GET")
	ar.HandleFunc("/search", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeCommentsSearch))
	ar.HandleFunc("/search", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeNewsSearch))
	ar.HandleFunc("/search", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeBlogSearch))
	ar.HandleFunc("/search", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeLinkerSearch))
	ar.HandleFunc("/search", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeWritingSearch))
	ar.HandleFunc("/search", adminSearchRemakeImageSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeImageSearch))
	ar.HandleFunc("/search/list", adminSearchWordListPage).Methods("GET")
	ar.HandleFunc("/search/list.txt", adminSearchWordListDownloadPage).Methods("GET")

	// forum admin routes
	far := ar.PathPrefix("/forum").Subrouter()
	far.HandleFunc("", adminForumPage).Methods("GET")
	far.HandleFunc("", adminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeStatisticInformationOnForumthread))
	far.HandleFunc("", adminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeStatisticInformationOnForumtopic))
	far.HandleFunc("/flagged", adminForumFlaggedPostsPage).Methods("GET")
	far.HandleFunc("/logs", adminForumModeratorLogsPage).Methods("GET")
	far.HandleFunc("/list", adminForumWordListPage).Methods("GET")
	far.HandleFunc("/categories", forumAdminCategoriesPage).Methods("GET")
	far.HandleFunc("/categories", common.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/category/{category}", forumAdminCategoryEditPage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumCategoryChange))
	far.HandleFunc("/category", forumAdminCategoryCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumCategoryCreate))
	far.HandleFunc("/category/delete", forumAdminCategoryDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	far.HandleFunc("/topics", forumAdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", common.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/conversations", forumAdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", forumAdminThreadDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumThreadDelete))
	far.HandleFunc("/topic/{topic}/edit", forumAdminTopicEditPage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicChange))
	far.HandleFunc("/topic/{topic}/delete", forumAdminTopicDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicDelete))
	far.HandleFunc("/topic", forumTopicCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicCreate))
	far.HandleFunc("/topic/{topic}/levels", forumAdminTopicRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/levels", forumAdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forumAdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskSetTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forumAdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forumAdminTopicRestrictionLevelCopyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCopyTopicRestriction))
	far.HandleFunc("/users", forumAdminUserPage).Methods("GET")
	far.HandleFunc("/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskSetUserLevel))
	far.HandleFunc("/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskUpdateUserLevel))
	far.HandleFunc("/user/{user}/levels", forumAdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher(TaskDeleteUserLevel))
	far.HandleFunc("/user/{user}/levels", forumAdminUserLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/users", forumAdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher(TaskDeleteUserLevel))
	far.HandleFunc("/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskUpdateUserLevel))
	far.HandleFunc("/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskSetUserLevel))
	far.HandleFunc("/restrictions/users", forumAdminUsersRestrictionsPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", forumAdminTopicsRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateTopicRestriction))
	far.HandleFunc("/restrictions/topics", forumAdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteTopicRestriction))
	far.HandleFunc("/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskSetTopicRestriction))
	far.HandleFunc("/restrictions/topics", forumAdminTopicsRestrictionLevelCopyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCopyTopicRestriction))

	// linker admin
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", linkerAdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", linkerAdminCategoriesUpdatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	lar.HandleFunc("/categories", linkerAdminCategoriesRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameCategory))
	lar.HandleFunc("/categories", linkerAdminCategoriesDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	lar.HandleFunc("/categories", linkerAdminCategoriesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateCategory))
	lar.HandleFunc("/add", linkerAdminAddPage).Methods("GET")
	lar.HandleFunc("/add", linkerAdminAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAdd))
	lar.HandleFunc("/queue", linkerAdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", linkerAdminQueueDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	lar.HandleFunc("/queue", linkerAdminQueueApproveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskApprove))
	lar.HandleFunc("/queue", linkerAdminQueueUpdateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	lar.HandleFunc("/queue", linkerAdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskBulkApprove))
	lar.HandleFunc("/queue", linkerAdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskBulkDelete))
	lar.HandleFunc("/users/levels", linkerAdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", linkerAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	lar.HandleFunc("/users/levels", linkerAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))

	// faq admin
	faq.RegisterAdminRoutes(ar)

	// languages
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages", adminLanguagesRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameLanguage))
	ar.HandleFunc("/languages", adminLanguagesDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteLanguage))
	ar.HandleFunc("/languages", adminLanguagesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateLanguage))

	// news admin
	nar := ar.PathPrefix("/news").Subrouter()
	nar.HandleFunc("/users/levels", newsAdminUserLevelsPage).Methods("GET")
	nar.HandleFunc("/users/levels", newsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAllow))
	nar.HandleFunc("/users/levels", newsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemoveLower))

	// writings admin
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", writingsUserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", writingsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	war.HandleFunc("/users/permissions", writingsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/users/levels", writingsAdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", writingsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	war.HandleFunc("/users/levels", writingsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/users/access", writingsAdminUserAccessPage).Methods("GET")
	war.HandleFunc("/users/access", writingsAdminUserAccessAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAddApproval))
	war.HandleFunc("/users/access", writingsAdminUserAccessUpdateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateUserApproval))
	war.HandleFunc("/users/access", writingsAdminUserAccessRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteUserApproval))
	war.HandleFunc("/category/{category}/permissions", writingsCategoryPermissionsPage).Methods("GET")
	war.HandleFunc("/category/{category}/permissions", writingsCategoryPermissionsAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	war.HandleFunc("/category/{category}/permissions/delete", writingsCategoryPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/categories", writingsAdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", writingsAdminCategoriesModifyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskWritingCategoryChange))
	war.HandleFunc("/categories", writingsAdminCategoriesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskWritingCategoryCreate))

	ar.HandleFunc("/reload", adminReloadConfigPage).Methods("POST")
	ar.HandleFunc("/shutdown", adminShutdownPage).Methods("POST")
}
