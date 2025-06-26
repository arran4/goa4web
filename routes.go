package goa4web

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
	"net/http"

	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	auth "github.com/arran4/goa4web/handlers/auth"
	blogs "github.com/arran4/goa4web/handlers/blogs"
	bookmarks "github.com/arran4/goa4web/handlers/bookmarks"
	comments "github.com/arran4/goa4web/handlers/comments"
	"github.com/arran4/goa4web/handlers/common"
	faq "github.com/arran4/goa4web/handlers/faq"
	forum "github.com/arran4/goa4web/handlers/forum"
	imagebbs "github.com/arran4/goa4web/handlers/imagebbs"
	languages "github.com/arran4/goa4web/handlers/languages"
	linker "github.com/arran4/goa4web/handlers/linker"
	news "github.com/arran4/goa4web/handlers/news"
	search "github.com/arran4/goa4web/handlers/search"
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
	nr.HandleFunc("/news/{post}", news.NewsPostReplyActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskReply))
	nr.HandleFunc("/news/{post}", news.NewsPostEditActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(common.TaskMatcher(TaskEdit))
	nr.HandleFunc("/news/{post}", news.NewsPostNewActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(common.TaskMatcher(TaskNewPost))
	nr.HandleFunc("/news/{post}/announcement", news.NewsAnnouncementActivateActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskAdd))
	nr.HandleFunc("/news/{post}/announcement", news.NewsAnnouncementDeactivateActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskDelete))
	nr.HandleFunc("/news/{post}", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCancel))
	nr.HandleFunc("/news/{post}", common.TaskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", news.NewsUserPermissionsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", news.NewsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher("User Allow"))
	nr.HandleFunc("/users/permissions", news.NewsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher("User Disallow"))
}

func registerBlogsRoutes(r *mux.Router) {
	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", blogs.RssPage).Methods("GET")
	br.HandleFunc("/atom", blogs.AtomPage).Methods("GET")
	br.HandleFunc("", blogs.Page).Methods("GET")
	br.HandleFunc("/", blogs.Page).Methods("GET")
	br.HandleFunc("/add", blogs.BlogAddPage).Methods("GET").MatcherFunc(auth.RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", blogs.BlogAddActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("writer", "administrator")).MatcherFunc(common.TaskMatcher(TaskAdd))
	br.HandleFunc("/bloggers", blogs.BloggersPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", blogs.BloggerPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", blogs.BloggerPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", blogs.BlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", common.TaskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", blogs.CommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", blogs.BlogReplyPostPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogs.CommentEditPostPage).MatcherFunc(Or(auth.RequiredAccess("administrator"), comments.Author())).Methods("POST").MatcherFunc(common.TaskMatcher(TaskEditReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogs.CommentEditPostCancelPage).MatcherFunc(Or(auth.RequiredAccess("administrator"), comments.Author())).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCancel))
	br.HandleFunc("/blog/{blog}/edit", blogs.BlogEditPage).Methods("GET").MatcherFunc(Or(auth.RequiredAccess("administrator"), And(auth.RequiredAccess("writer"), blogs.BlogAuthor())))
	br.HandleFunc("/blog/{blog}/edit", blogs.BlogEditActionPage).Methods("POST").MatcherFunc(Or(auth.RequiredAccess("administrator"), And(auth.RequiredAccess("writer"), blogs.BlogAuthor()))).MatcherFunc(common.TaskMatcher(TaskEdit))
	br.HandleFunc("/blog/{blog}/edit", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCancel))

	// Admin endpoints for blogs
	br.HandleFunc("/user/permissions", blogs.GetPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskUserAllow))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskUserDisallow))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskUsersAllow))
	br.HandleFunc("/users/permissions", blogs.UsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskUsersDisallow))
}

func registerForumRoutes(r *mux.Router) {
	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("/topic/{topic}.rss", forum.TopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", forum.TopicAtomPage).Methods("GET")
	fr.HandleFunc("", forum.Page).Methods("GET")
	fr.HandleFunc("/category/{category}", forum.Page).Methods("GET")
	fr.HandleFunc("/topic/{topic}", forum.TopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forum.ThreadNewPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forum.ThreadNewActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCreateThread))
	fr.HandleFunc("/topic/{topic}/thread", forum.ThreadNewCancelPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}", forum.ThreadPage).Methods("GET").MatcherFunc(forum.GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(forum.GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forum.TopicThreadReplyPage).Methods("POST").MatcherFunc(forum.GetThreadAndTopic()).MatcherFunc(common.TaskMatcher(TaskReply))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forum.TopicThreadReplyCancelPage).Methods("POST").MatcherFunc(forum.GetThreadAndTopic()).MatcherFunc(common.TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forum.TopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(forum.GetThreadAndTopic()).MatcherFunc(common.TaskMatcher(TaskEditReply)).MatcherFunc(Or(auth.RequiredAccess("administrator"), comments.Author()))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forum.TopicThreadCommentEditActionCancelPage).Methods("POST").MatcherFunc(forum.GetThreadAndTopic()).MatcherFunc(common.TaskMatcher(TaskCancel))
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
	lr.HandleFunc("/comments/{link}", linker.CommentsReplyPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskReply))
	lr.HandleFunc("/show/{link}", linker.ShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", linker.ShowReplyPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskReply))
	lr.HandleFunc("/suggest", linker.SuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", linker.SuggestActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSuggest))
}

func registerBookmarksRoutes(r *mux.Router) {
	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarks.Page).Methods("GET")
	bmr.HandleFunc("/mine", bookmarks.MinePage).Methods("GET").MatcherFunc(auth.RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarks.EditPage).Methods("GET").MatcherFunc(auth.RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarks.EditSaveActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskSave))
	bmr.HandleFunc("/edit", bookmarks.EditCreateActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskCreate))
	bmr.HandleFunc("/edit", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount())
}

func registerImagebbsRoutes(r *mux.Router) {
	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.PathPrefix("/images/").Handler(http.StripPrefix("/imagebbs/images/", http.FileServer(http.Dir(runtimeconfig.AppRuntimeConfig.ImageUploadDir))))
	ibr.HandleFunc(".rss", imagebbs.RssPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", imagebbs.BoardRssPage).Methods("GET")
	ibr.HandleFunc(".atom", imagebbs.AtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", imagebbs.BoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbs.BoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbs.BoardPostImageActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskUploadImage))
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbs.BoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbs.BoardThreadReplyActionPage).Methods("POST").MatcherFunc(auth.RequiresAnAccount()).MatcherFunc(common.TaskMatcher(TaskReply))
	ibr.HandleFunc("", imagebbs.Page).Methods("GET")
	ibr.HandleFunc("/", imagebbs.Page).Methods("GET")
	ibr.HandleFunc("/poster/{username}", imagebbs.PosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", imagebbs.PosterPage).Methods("GET")

	// Admin endpoints for image boards
	ibr.HandleFunc("/admin", imagebbs.AdminPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", imagebbs.AdminBoardsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbs.AdminNewBoardPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbs.AdminNewBoardMakePage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskNewBoard))
	ibr.HandleFunc("/admin/board", common.TaskDoneAutoRefreshPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board/{board}", imagebbs.AdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskModifyBoard))
	ibr.HandleFunc("/admin/approve/{post}", imagebbs.AdminApprovePostPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskApprove))
	ibr.HandleFunc("/admin/files", imagebbs.AdminFilesPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
}

func registerSearchRoutes(r *mux.Router) {
	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", search.Page).Methods("GET")
	sr.HandleFunc("", search.SearchResultForumActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSearchForum))
	sr.HandleFunc("", news.SearchResultNewsActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSearchNews))
	sr.HandleFunc("", search.SearchResultLinkerActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSearchLinker))
	sr.HandleFunc("", search.SearchResultBlogsActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSearchBlogs))
	sr.HandleFunc("", search.SearchResultWritingsActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSearchWritings))
}

func registerWritingsRoutes(r *mux.Router) {
	wr := r.PathPrefix("/writings").Subrouter()
	wr.HandleFunc("/rss", writings.RssPage).Methods("GET")
	wr.HandleFunc("/atom", writings.AtomPage).Methods("GET")
	wr.HandleFunc("", writings.Page).Methods("GET")
	wr.HandleFunc("/", writings.Page).Methods("GET")
	wr.HandleFunc("/writer/{username}", writings.WriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", writings.WriterPage).Methods("GET")
	wr.HandleFunc("/user/permissions", writings.UserPermissionsPage).Methods("GET").MatcherFunc(auth.RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", writings.UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskUserAllow))
	wr.HandleFunc("/users/permissions", writings.UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(auth.RequiredAccess("administrator")).MatcherFunc(common.TaskMatcher(TaskUserDisallow))
	wr.HandleFunc("/article/{article}", writings.ArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", writings.ArticleReplyActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskReply))
	wr.HandleFunc("/article/{article}/edit", writings.ArticleEditPage).Methods("GET").MatcherFunc(Or(And(auth.RequiredAccess("writer"), writings.WritingAuthor()), auth.RequiredAccess("administrator")))
	wr.HandleFunc("/article/{article}/edit", writings.ArticleEditActionPage).Methods("POST").MatcherFunc(Or(And(auth.RequiredAccess("writer"), writings.WritingAuthor()), auth.RequiredAccess("administrator"))).MatcherFunc(common.TaskMatcher(TaskUpdateWriting))
	wr.HandleFunc("/categories", writings.CategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", writings.CategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", writings.CategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", writings.ArticleAddPage).Methods("GET").MatcherFunc(Or(auth.RequiredAccess("writer"), auth.RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", writings.ArticleAddActionPage).Methods("POST").MatcherFunc(Or(auth.RequiredAccess("writer"), auth.RequiredAccess("administrator"))).MatcherFunc(common.TaskMatcher(TaskSubmitWriting))
}

func registerWritingsAdminRoutes(ar *mux.Router) {
	war := ar.PathPrefix("/writings").Subrouter()
	war.HandleFunc("/user/permissions", writings.UserPermissionsPage).Methods("GET")
	war.HandleFunc("/users/permissions", writings.UsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserAllow))
	war.HandleFunc("/users/permissions", writings.UsersPermissionsDisallowPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/users/levels", writings.AdminUserLevelsPage).Methods("GET")
	war.HandleFunc("/users/levels", writings.AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserAllow))
	war.HandleFunc("/users/levels", writings.AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/users/access", writings.AdminUserAccessPage).Methods("GET")
	war.HandleFunc("/users/access", writings.AdminUserAccessAddActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskAddApproval))
	war.HandleFunc("/users/access", writings.AdminUserAccessUpdateActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUpdateUserApproval))
	war.HandleFunc("/users/access", writings.AdminUserAccessRemoveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDeleteUserApproval))
	war.HandleFunc("/category/{category}/permissions", writings.CategoryPermissionsPage).Methods("GET")
	war.HandleFunc("/category/{category}/permissions", writings.CategoryPermissionsAllowPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserAllow))
	war.HandleFunc("/category/{category}/permissions/delete", writings.CategoryPermissionsDisallowPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserDisallow))
	war.HandleFunc("/categories", writings.AdminCategoriesPage).Methods("GET")
	war.HandleFunc("/categories", writings.AdminCategoriesModifyPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskWritingCategoryChange))
	war.HandleFunc("/categories", writings.AdminCategoriesCreatePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskWritingCategoryCreate))
}

func registerInformationRoutes(r *mux.Router) {
	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", informationPage).Methods("GET")
}

func registerRegisterRoutes(r *mux.Router) {
	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", auth.RegisterPage).Methods("GET").MatcherFunc(Not(auth.RequiresAnAccount()))
	rr.HandleFunc("", auth.RegisterActionPage).Methods("POST").MatcherFunc(Not(auth.RequiresAnAccount())).MatcherFunc(common.TaskMatcher(TaskRegister))
}

func registerLoginRoutes(r *mux.Router) {
	ulr := r.PathPrefix("/login").Subrouter()
	ulr.HandleFunc("", auth.LoginUserPassPage).Methods("GET").MatcherFunc(Not(auth.RequiresAnAccount()))
	ulr.HandleFunc("", auth.LoginActionPage).Methods("POST").MatcherFunc(Not(auth.RequiresAnAccount())).MatcherFunc(common.TaskMatcher(TaskLogin))
}

func registerAdminRoutes(r *mux.Router) {
	ar := r.PathPrefix("/admin").Subrouter()
	ar.Use(AdminCheckerMiddleware)
	ar.HandleFunc("", adminhandlers.AdminPage).Methods("GET")
	ar.HandleFunc("/", adminhandlers.AdminPage).Methods("GET")
	ar.HandleFunc("/categories", adminhandlers.AdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/permissions/sections", adminhandlers.AdminPermissionsSectionPage).Methods("GET")
	ar.HandleFunc("/permissions/sections/view", adminhandlers.AdminPermissionsSectionViewPage).Methods("GET")
	ar.HandleFunc("/permissions/sections", adminhandlers.AdminPermissionsSectionRenamePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskRenameSection))
	ar.HandleFunc("/email/queue", adminhandlers.AdminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/queue", adminhandlers.AdminEmailQueueResendActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskResend))
	ar.HandleFunc("/email/queue", adminhandlers.AdminEmailQueueDeleteActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDelete))
	ar.HandleFunc("/email/template", adminhandlers.AdminEmailTemplatePage).Methods("GET")
	ar.HandleFunc("/email/template", adminhandlers.AdminEmailTemplateSaveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUpdate))
	ar.HandleFunc("/email/template", adminhandlers.AdminEmailTemplateTestActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskTestMail))
	ar.HandleFunc("/notifications", adminhandlers.AdminNotificationsPage).Methods("GET")
	ar.HandleFunc("/notifications", adminhandlers.AdminNotificationsMarkReadActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDismiss))
	ar.HandleFunc("/notifications", adminhandlers.AdminNotificationsPurgeActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskPurge))
	ar.HandleFunc("/notifications", adminhandlers.AdminNotificationsSendActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskNotify))
	ar.HandleFunc("/announcements", adminhandlers.AdminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", adminhandlers.AdminAnnouncementsAddActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskAdd))
	ar.HandleFunc("/announcements", adminhandlers.AdminAnnouncementsDeleteActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDelete))
	ar.HandleFunc("/ipbans", adminhandlers.AdminIPBanPage).Methods("GET")
	ar.HandleFunc("/ipbans", adminhandlers.AdminIPBanAddActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskAdd))
	ar.HandleFunc("/ipbans", adminhandlers.AdminIPBanDeleteActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDelete))
	ar.HandleFunc("/audit", adminhandlers.AdminAuditLogPage).Methods("GET")
	ar.HandleFunc("/settings", adminhandlers.AdminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/stats", adminhandlers.AdminServerStatsPage).Methods("GET")
	ar.HandleFunc("/usage", adminhandlers.AdminUsageStatsPage).Methods("GET")

	// search related

	// forum admin routes
	far := ar.PathPrefix("/forum").Subrouter()
	far.HandleFunc("", forum.AdminForumPage).Methods("GET")
	far.HandleFunc("", forum.AdminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskRemakeStatisticInformationOnForumthread))
	far.HandleFunc("", forum.AdminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskRemakeStatisticInformationOnForumtopic))
	far.HandleFunc("/flagged", forum.AdminForumFlaggedPostsPage).Methods("GET")
	far.HandleFunc("/logs", forum.AdminForumModeratorLogsPage).Methods("GET")
	far.HandleFunc("/list", forum.AdminForumWordListPage).Methods("GET")
	far.HandleFunc("/categories", forum.AdminCategoriesPage).Methods("GET")
	far.HandleFunc("/categories", common.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/category/{category}", forum.AdminCategoryEditPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskForumCategoryChange))
	far.HandleFunc("/category", forum.AdminCategoryCreatePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskForumCategoryCreate))
	far.HandleFunc("/category/delete", forum.AdminCategoryDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDeleteCategory))
	far.HandleFunc("/topics", forum.AdminTopicsPage).Methods("GET")
	far.HandleFunc("/topics", common.TaskDoneAutoRefreshPage).Methods("POST")
	far.HandleFunc("/conversations", forum.AdminThreadsPage).Methods("GET")
	far.HandleFunc("/thread/{thread}/delete", forum.AdminThreadDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskForumThreadDelete))
	far.HandleFunc("/topic/{topic}/edit", forum.AdminTopicEditPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskForumTopicChange))
	far.HandleFunc("/topic/{topic}/delete", forum.AdminTopicDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskForumTopicDelete))
	far.HandleFunc("/topic", forum.TopicCreatePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskForumTopicCreate))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUpdateTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSetTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDeleteTopicRestriction))
	far.HandleFunc("/topic/{topic}/levels", forum.AdminTopicRestrictionLevelCopyPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCopyTopicRestriction))
	far.HandleFunc("/users", forum.AdminUserPage).Methods("GET")
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(forum.AdminUsersMaxLevelNotLowerThanTargetLevel(), forum.TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(common.TaskMatcher(TaskSetUserLevel))
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(forum.AdminUsersMaxLevelNotLowerThanTargetLevel(), forum.TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(common.TaskMatcher(TaskUpdateUserLevel))
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(forum.AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(common.TaskMatcher(TaskDeleteUserLevel))
	far.HandleFunc("/user/{user}/levels", forum.AdminUserLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(forum.AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(common.TaskMatcher(TaskDeleteUserLevel))
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(forum.AdminUsersMaxLevelNotLowerThanTargetLevel(), forum.TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(common.TaskMatcher(TaskUpdateUserLevel))
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(forum.AdminUsersMaxLevelNotLowerThanTargetLevel(), forum.TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(common.TaskMatcher(TaskSetUserLevel))
	far.HandleFunc("/restrictions/users", forum.AdminUsersRestrictionsPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelPage).Methods("GET")
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUpdateTopicRestriction))
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDeleteTopicRestriction))
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskSetTopicRestriction))
	far.HandleFunc("/restrictions/topics", forum.AdminTopicsRestrictionLevelCopyPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCopyTopicRestriction))

	// linker admin
	lar := ar.PathPrefix("/linker").Subrouter()
	lar.HandleFunc("/categories", linker.AdminCategoriesPage).Methods("GET")
	lar.HandleFunc("/categories", linker.AdminCategoriesUpdatePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUpdate))
	lar.HandleFunc("/categories", linker.AdminCategoriesRenamePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskRenameCategory))
	lar.HandleFunc("/categories", linker.AdminCategoriesDeletePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDeleteCategory))
	lar.HandleFunc("/categories", linker.AdminCategoriesCreatePage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskCreateCategory))
	lar.HandleFunc("/add", linker.AdminAddPage).Methods("GET")
	lar.HandleFunc("/add", linker.AdminAddActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskAdd))
	lar.HandleFunc("/queue", linker.AdminQueuePage).Methods("GET")
	lar.HandleFunc("/queue", linker.AdminQueueDeleteActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskDelete))
	lar.HandleFunc("/queue", linker.AdminQueueApproveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskApprove))
	lar.HandleFunc("/queue", linker.AdminQueueUpdateActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUpdate))
	lar.HandleFunc("/queue", linker.AdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskBulkApprove))
	lar.HandleFunc("/queue", linker.AdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskBulkDelete))
	lar.HandleFunc("/users/levels", linker.AdminUserLevelsPage).Methods("GET")
	lar.HandleFunc("/users/levels", linker.AdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserAllow))
	lar.HandleFunc("/users/levels", linker.AdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskUserDisallow))

	// faq admin
	faq.RegisterAdminRoutes(ar)
	search.RegisterAdminRoutes(ar)
	userhandlers.RegisterAdminRoutes(ar)
	languages.RegisterAdminRoutes(ar)

	// news admin
	nar := ar.PathPrefix("/news").Subrouter()
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsPage).Methods("GET")
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskAllow))
	nar.HandleFunc("/users/levels", news.NewsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(common.TaskMatcher(TaskRemoveLower))

	// writings admin
	registerWritingsAdminRoutes(ar)

	ar.HandleFunc("/reload", adminhandlers.AdminReloadConfigPage).Methods("POST")
	ar.HandleFunc("/shutdown", adminhandlers.AdminShutdownPage).Methods("POST")
}
