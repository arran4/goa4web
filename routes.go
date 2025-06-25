package goa4web

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
	"net/http"

	blogs "github.com/arran4/goa4web/handlers/blogs"
	bookmarks "github.com/arran4/goa4web/handlers/bookmarks"
	"github.com/arran4/goa4web/handlers/common"
	faq "github.com/arran4/goa4web/handlers/faq"
	forum "github.com/arran4/goa4web/handlers/forum"
	imagebbs "github.com/arran4/goa4web/handlers/imagebbs"
	linker "github.com/arran4/goa4web/handlers/linker"
	news "github.com/arran4/goa4web/handlers/news"
	writings "github.com/arran4/goa4web/handlers/writings"

	userhandlers "github.com/arran4/goa4web/handlers/user"
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
	userhandlers.RegisterRoutes(r)
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
	nr.HandleFunc(".rss", news.NewsRssPage).Methods("GET")
	nr.HandleFunc("", runTemplate("page.gohtml")).Methods("GET")
	nr.HandleFunc("", common.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", news.NewsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", news.NewsPostReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskReply))
	nr.HandleFunc("/news/{post}", news.NewsPostEditActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskEdit))
	nr.HandleFunc("/news/{post}", news.NewsPostNewActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskNewPost))
	nr.HandleFunc("/news/{post}/announcement", news.NewsAnnouncementActivateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskAdd))
	nr.HandleFunc("/news/{post}/announcement", news.NewsAnnouncementDeactivateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskDelete))
	nr.HandleFunc("/news/{post}", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	nr.HandleFunc("/news/{post}", common.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", news.NewsUserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", news.NewsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Allow"))
	nr.HandleFunc("/users/permissions", news.NewsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Disallow"))
}

func registerBlogsRoutes(r *mux.Router) {
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", blogs.RssPage).Methods("GET")
	br.HandleFunc("/atom", blogs.AtomPage).Methods("GET")
	br.HandleFunc("", blogs.Page).Methods("GET")
	br.HandleFunc("/", blogs.Page).Methods("GET")
	br.HandleFunc("/add", blogs.BlogAddPage).Methods("GET").MatcherFunc(RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", blogs.BlogAddActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskAdd))
	br.HandleFunc("/bloggers", blogs.BloggersPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", blogs.BloggerPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", blogs.BloggerPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", blogs.BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", common.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", blogs.CommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", blogs.BlogReplyPostPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogs.CommentEditPostPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher(TaskEditReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogs.CommentEditPostCancelPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	br.HandleFunc("/blog/{blog}/edit", blogs.BlogEditPage).Methods("GET").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor())))
	br.HandleFunc("/blog/{blog}/edit", blogs.BlogEditActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor()))).MatcherFunc(TaskMatcher(TaskEdit))
	br.HandleFunc("/blog/{blog}/edit", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", blogs.GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserAllow))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserDisallow))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUsersAllow))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUsersDisallow))
}

func registerForumRoutes(r *mux.Router) {
	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("/topic/{topic}.rss", forum.TopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", forum.TopicAtomPage).Methods("GET")
	fr.HandleFunc("", forum.Page).Methods("GET")
	fr.HandleFunc("/category/{category}", forum.Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}", forum.TopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forum.ThreadNewPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forum.ThreadNewActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateThread))
	fr.HandleFunc("/topic/{topic}/thread", forum.ThreadNewCancelPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}", forum.ThreadPage).Methods("GET").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forum.TopicThreadReplyPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskReply))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forum.TopicThreadReplyCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forum.TopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskEditReply)).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor()))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forum.TopicThreadCommentEditActionCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskCancel))
}

func registerLinkerRoutes(r *mux.Router) {
	lr := r.PathPrefix("/linker").Subrouter()
	lr.HandleFunc("/rss", linker.RssPage).Methods("GET")
	lr.HandleFunc("/atom", linker.AtomPage).Methods("GET")
	lr.HandleFunc("", linker.Page).Methods("GET")
	lr.HandleFunc("/linker/{username}", linker.LinkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", linker.LinkerPage).Methods("GET")
	lr.HandleFunc("/categories", linker.CategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", linker.CategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linker.CommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linker.CommentsReplyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	lr.HandleFunc("/show/{link}", linker.ShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", linker.ShowReplyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	lr.HandleFunc("/suggest", linker.SuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", linker.SuggestActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSuggest))
}

func registerBookmarksRoutes(r *mux.Router) {
	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarks.Page).Methods("GET")
	bmr.HandleFunc("/mine", bookmarks.MinePage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarks.EditPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarks.EditSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSave))
	bmr.HandleFunc("/edit", bookmarks.EditCreateActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskCreate))
	bmr.HandleFunc("/edit", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiresAnAccount())
}

func registerImagebbsRoutes(r *mux.Router) {
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(runtimeconfig.AppRuntimeConfig.ImageUploadDir))))
	ibr.HandleFunc(".rss", imagebbs.RssPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", imagebbs.BoardRssPage).Methods("GET")
	ibr.HandleFunc(".atom", imagebbs.AtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", imagebbs.BoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbs.BoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbs.BoardPostImageActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskUploadImage))
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbs.BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbs.BoardThreadReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskReply))
	ibr.HandleFunc("", imagebbs.Page).Methods("GET")
	ibr.HandleFunc("/", imagebbs.Page).Methods("GET")
	ibr.HandleFunc("/poster/{username}", imagebbs.PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", imagebbs.PosterPage).Methods("GET")

	// Admin endpoints for image boards
	ibr.HandleFunc("/admin", imagebbs.AdminPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", imagebbs.AdminBoardsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbs.AdminNewBoardPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbs.AdminNewBoardMakePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskNewBoard))
	ibr.HandleFunc("/admin/board", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board/{board}", imagebbs.AdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskModifyBoard))
	ibr.HandleFunc("/admin/approve/{post}", imagebbs.AdminApprovePostPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskApprove))
	ibr.HandleFunc("/admin/files", imagebbs.AdminFilesPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
}

func registerSearchRoutes(r *mux.Router) {
	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", searchPage).Methods("GET")
	sr.HandleFunc("", searchResultForumActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchForum))
	sr.HandleFunc("", news.SearchResultNewsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchNews))
	sr.HandleFunc("", searchResultLinkerActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchLinker))
	sr.HandleFunc("", searchResultBlogsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchBlogs))
	sr.HandleFunc("", searchResultWritingsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchWritings))
}

func registerWritingsRoutes(r *mux.Router) {
	wr := r.PathPrefix("/writings").Subrouter()
	wr.HandleFunc("/rss", writings.RssPage).Methods("GET")
	wr.HandleFunc("/atom", writings.AtomPage).Methods("GET")
	wr.HandleFunc("", writings.Page).Methods("GET")
	wr.HandleFunc("/", writings.Page).Methods("GET")
	wr.HandleFunc("/writer/{username}", writings.WriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", writings.WriterPage).Methods("GET")
	wr.HandleFunc("/user/permissions", writings.UserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", writings.UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserAllow))
	wr.HandleFunc("/users/permissions", writings.UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserDisallow))
	wr.HandleFunc("/article/{article}", writings.ArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", writings.ArticleReplyActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	wr.HandleFunc("/article/{article}/edit", writings.ArticleEditPage).Methods("GET").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator")))
	wr.HandleFunc("/article/{article}/edit", writings.ArticleEditActionPage).Methods("POST").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher(TaskUpdateWriting))
	wr.HandleFunc("/categories", writings.CategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", writings.CategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", writings.CategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", writings.ArticleAddPage).Methods("GET").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", writings.ArticleAddActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher(TaskSubmitWriting))
}

func registerWritingsAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", writings.UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", writings.UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	war.HandleFunc("/users/permissions", writings.UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/users/levels", writings.AdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", writings.AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	war.HandleFunc("/users/levels", writings.AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/users/access", writings.AdminUserAccessPage).Methods("GET")
	war.HandleFunc("/users/access", writings.AdminUserAccessAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAddApproval))
	war.HandleFunc("/users/access", writings.AdminUserAccessUpdateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateUserApproval))
	war.HandleFunc("/users/access", writings.AdminUserAccessRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteUserApproval))
	war.HandleFunc("/category/{category}/permissions", writings.CategoryPermissionsPage).Methods("GET")
	war.HandleFunc("/category/{category}/permissions", writings.CategoryPermissionsAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	war.HandleFunc("/category/{category}/permissions/delete", writings.CategoryPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/categories", writings.AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", writings.AdminCategoriesModifyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskWritingCategoryChange))
	war.HandleFunc("/categories", writings.AdminCategoriesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskWritingCategoryCreate))
}

func registerInformationRoutes(r *mux.Router) {
	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", informationPage).Methods("GET")
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
	far.HandleFunc("", forum.AdminForumPage).Methods("GET")
	far.HandleFunc("", forum.AdminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeStatisticInformationOnForumthread))
	far.HandleFunc("", forum.AdminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeStatisticInformationOnForumtopic))
	far.HandleFunc("/flagged", forum.AdminForumFlaggedPostsPage).Methods("GET")
	far.HandleFunc("/logs", forum.AdminForumModeratorLogsPage).Methods("GET")
	far.HandleFunc("/list", forum.AdminForumWordListPage).Methods("GET")
	far.HandleFunc("/categories", forum.AdminCategoriesPage).Methods("GET")
	far.HandleFunc("/categories", common.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/category/{category}", forum.AdminCategoryEditPage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumCategoryChange))
	far.HandleFunc("/category", forum.AdminCategoryCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumCategoryCreate))
	far.HandleFunc("/category/delete", forum.AdminCategoryDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	far.HandleFunc("/topics", forum.AdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", common.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/conversations", forum.AdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", forum.AdminThreadDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumThreadDelete))
	far.HandleFunc("/topic/{topic}/edit", forum.AdminTopicEditPage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicChange))
	far.HandleFunc("/topic/{topic}/delete", forum.AdminTopicDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicDelete))
	far.HandleFunc("/topic", forum.TopicCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicCreate))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskSetTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelCopyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCopyTopicRestriction))
	far.HandleFunc("/users", forum.AdminUserPage).Methods("GET")
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskSetUserLevel))
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskUpdateUserLevel))
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher(TaskDeleteUserLevel))
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher(TaskDeleteUserLevel))
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskUpdateUserLevel))
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskSetUserLevel))
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateTopicRestriction))
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteTopicRestriction))
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskSetTopicRestriction))
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelCopyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCopyTopicRestriction))

	// linker admin
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", linker.AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", linker.AdminCategoriesUpdatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	lar.HandleFunc("/categories", linker.AdminCategoriesRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameCategory))
	lar.HandleFunc("/categories", linker.AdminCategoriesDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	lar.HandleFunc("/categories", linker.AdminCategoriesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateCategory))
	lar.HandleFunc("/add", linker.AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", linker.AdminAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAdd))
	lar.HandleFunc("/queue", linker.AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", linker.AdminQueueDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	lar.HandleFunc("/queue", linker.AdminQueueApproveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskApprove))
	lar.HandleFunc("/queue", linker.AdminQueueUpdateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	lar.HandleFunc("/queue", linker.AdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskBulkApprove))
	lar.HandleFunc("/queue", linker.AdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskBulkDelete))
	lar.HandleFunc("/users/levels", linker.AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", linker.AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	lar.HandleFunc("/users/levels", linker.AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))

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
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsPage).Methods("GET")
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAllow))
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemoveLower))

	// writings admin
	registerWritingsAdminRoutes(ar)

	ar.HandleFunc("/reload", adminReloadConfigPage).Methods("POST")
	ar.HandleFunc("/shutdown", adminShutdownPage).Methods("POST")
}
