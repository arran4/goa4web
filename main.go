package main

import (
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	//	// Replace these with your Google OAuth2 credentials
	//	clientID     = ""
	//	clientSecret = ""
	//	redirectURL  = "http://localhost:8080/callback"
	//
	//	// Change this to your desired session key
	sessionName = "my-session"
	//sessionKey  = "authenticated"
	store = sessions.NewCookieStore([]byte("random-key"))
	//
	//	oauth2Config = oauth2.Config{
	//		ClientID:     clientID,
	//		ClientSecret: clientSecret,
	//		RedirectURL:  redirectURL,
	//		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	//		Endpoint:     endpoints.Google,
	//	}
)

func NewFuncs() template.FuncMap {
	return map[string]any{
		//"getSecurityLevel":
		"now": func() time.Time { return time.Now() },
		"a4code2html": func(s string) template.HTML {
			c := NewA4Code2HTML()
			c.codeType = ct_html
			c.input = s
			c.Process()
			return template.HTML(c.output.String())
		},
		"a4code2string": func(s string) string {
			c := NewA4Code2HTML()
			c.codeType = ct_wordsonly
			c.input = s
			c.Process()
			return c.output.String()
		},
		"firstline": func(s string) string {
			return strings.Split(s, "\n")[0]
		},
		"left": func(i int, s string) string {
			l := len(s)
			if l > i {
				l = i
			}
			return s[:l]
		},
	}
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	r := mux.NewRouter()

	r.Use(DBAdderMiddleware)
	r.Use(UserAdderMiddleware)
	r.Use(CoreAdderMiddleware)

	// TODO consider adsense / adwords / etc

	r.HandleFunc("/main.css", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(getMainCSSData())
	}).Methods("GET")

	// News
	r.HandleFunc("/", newsPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	//TODO nr.HandleFunc(".rss", newsRssPage).Methods("GET")
	nr.HandleFunc("", newsPage).Methods("GET")
	nr.HandleFunc("", taskDoneAutoRefreshPage).Methods("POST")
	//TODO nr.HandleFunc("/news/{id:[0-9]+}", newsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", newsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", taskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/news/{post}", newsPostReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Reply"))
	nr.HandleFunc("/news/{post}", newsPostEditActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher("Edit"))
	nr.HandleFunc("/news/{post}", newsPostNewActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher("New Post"))
	nr.HandleFunc("/news/admin/users/levels", newsAdminUserLevelsPage).Methods("GET")
	nr.HandleFunc("/news/admin/users/levels", newsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("allow"))
	nr.HandleFunc("/news/admin/users/levels", newsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("remove"))

	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.HandleFunc("", faqPage).Methods("GET", "POST")
	faqr.HandleFunc("/ask", faqAskPage).Methods("GET")
	faqr.HandleFunc("/ask", faqAskActionPage).Methods("POST").MatcherFunc(TaskMatcher("Ask"))
	faqr.HandleFunc("/admin/answer", faqAdminAnswerPage).Methods("GET", "POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(NoTask())
	faqr.HandleFunc("/admin/answer", faqAnswerAnswerActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Answer"))
	faqr.HandleFunc("/admin/answer", faqAnswerRemoveActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Remove"))
	faqr.HandleFunc("/admin/categories", faqAdminCategoriesPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	faqr.HandleFunc("/admin/categories", faqCategoriesRenameActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Rename Category"))
	faqr.HandleFunc("/admin/categories", faqCategoriesDeleteActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete Category"))
	faqr.HandleFunc("/admin/categories", faqCategoriesCreateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Create Category"))
	faqr.HandleFunc("/admin/questions", faqAdminQuestionsPage).Methods("GET", "POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(NoTask())
	faqr.HandleFunc("/admin/questions", faqQuestionsEditActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Edit"))
	faqr.HandleFunc("/admin/questions", faqQuestionsDeleteActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Remove"))

	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", blogsRssPage).Methods("GET")
	br.HandleFunc("/atom", blogsAtomPage).Methods("GET")
	br.HandleFunc("", blogsPage).Methods("GET")
	br.HandleFunc("/user/permissions", blogsUserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	br.HandleFunc("/users/permissions", blogsUsersPermissionsUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Allow"))
	br.HandleFunc("/users/permissions", blogsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Disallow"))
	br.HandleFunc("/add", blogsBlogAddPage).Methods("GET").MatcherFunc(RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", blogsBlogAddActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher("Add"))
	br.HandleFunc("/bloggers", blogsBloggersPage).Methods("GET")
	br.HandleFunc("/blogger/{blogger}", blogsBloggerPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", blogsBlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", taskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", blogsCommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", blogsBlogReplyPostPage).Methods("POST").MatcherFunc(TaskMatcher("Reply"))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogsCommentEditPostPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher("Edit Reply"))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogsCommentEditPostCancelPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher("Cancel"))
	br.HandleFunc("/blog/{blog}/edit", blogsBlogEditPage).Methods("GET").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor())))
	br.HandleFunc("/blog/{blog}/edit", blogsBlogEditActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor()))).MatcherFunc(TaskMatcher("Edit"))
	br.HandleFunc("/blog/{blog}/edit", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher("Cancel"))

	// TODO a matcher check to ensure topics and threads align.
	fr := r.PathPrefix("/forum").Subrouter()
	// TODO RSS & ATOM
	fr.HandleFunc("", forumPage).Methods("GET")
	fr.HandleFunc("/category/{category}", forumPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}", forumTopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewActionPage).Methods("POST").MatcherFunc(TaskMatcher("Create Thread"))
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewCancelPage).Methods("POST").MatcherFunc(TaskMatcher("Cancel"))
	fr.HandleFunc("/topic/{topic}/thread/{thread}", forumThreadPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread/{thread}", taskDoneAutoRefreshPage).Methods("POST")
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyPage).Methods("POST").MatcherFunc(TaskMatcher("Reply"))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyCancelPage).Methods("POST").MatcherFunc(TaskMatcher("Cancel"))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forumTopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(TaskMatcher("Edit Reply")).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor()))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forumTopicThreadCommentEditActionCancelPage).Methods("POST").MatcherFunc(TaskMatcher("Cancel"))
	fr.HandleFunc("/admin", forumAdminPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/categories", forumAdminCategoriesPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/categories", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/category/{category}", forumAdminCategoryEditPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Forum category change"))
	fr.HandleFunc("/admin/category", forumAdminCategoryCreatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Forum category create"))
	fr.HandleFunc("/admin/topics", forumAdminTopicsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/topics", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/topic/{topic}/edit", forumAdminTopicEditPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Forum topic change"))
	fr.HandleFunc("/admin/topic", forumTopicCreatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Forum topic create"))
	fr.HandleFunc("/admin/topic/{topic}/levels", forumAdminTopicRestrictionLevelPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/topic/{topic}/levels", forumAdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update topic restriction"))
	fr.HandleFunc("/admin/topic/{topic}/levels", forumAdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Set topic restriction"))
	fr.HandleFunc("/admin/topic/{topic}/levels", forumAdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete topic restriction"))
	fr.HandleFunc("/admin/users", forumAdminUserPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(RequiredAccess("administrator"), AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher("Set user level"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(RequiredAccess("administrator"), AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher("Update user level"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(RequiredAccess("administrator"), AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher("Delete user level"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(RequiredAccess("administrator"), AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher("Delete user level"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(RequiredAccess("administrator"), AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher("Update user level"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(RequiredAccess("administrator"), AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher("Set user level"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update topic restriction"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete topic restriction"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Set topic restriction"))

	lr := r.PathPrefix("/linker").Subrouter()
	//lr.HandleFunc(".rss", linkerRssPage).Methods("GET")
	//lr.HandleFunc(".atom", linkerAtomPage).Methods("GET")
	lr.HandleFunc("", linkerPage).Methods("GET")
	lr.HandleFunc("/categories", linkerCategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", linkerCategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linkerCommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linkerCommentsReplyPage).Methods("POST").MatcherFunc(TaskMatcher("Reply"))
	lr.HandleFunc("/show/{link}", linkerShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", linkerShowReplyPage).Methods("POST").MatcherFunc(TaskMatcher("Reply"))
	lr.HandleFunc("/suggest", linkerSuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", linkerSuggestActionPage).Methods("POST").MatcherFunc(TaskMatcher("Suggest"))
	lr.HandleFunc("/admin/categories", linkerAdminCategoriesPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	lr.HandleFunc("/admin/categories", linkerAdminCategoriesUpdatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update"))
	lr.HandleFunc("/admin/categories", linkerAdminCategoriesRenamePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Rename Category"))
	lr.HandleFunc("/admin/categories", linkerAdminCategoriesDeletePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete Category"))
	lr.HandleFunc("/admin/categories", linkerAdminCategoriesCreatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Create Category"))
	lr.HandleFunc("/admin/add", linkerAdminAddPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	lr.HandleFunc("/admin/add", linkerAdminAddActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Add"))
	lr.HandleFunc("/admin/queue", linkerAdminQueuePage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	lr.HandleFunc("/admin/queue", linkerAdminQueueDeleteActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete"))
	lr.HandleFunc("/admin/queue", linkerAdminQueueApproveActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Approve"))
	lr.HandleFunc("/admin/queue", linkerAdminQueueUpdateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update"))
	lr.HandleFunc("/admin/users/levels", linkerAdminUserLevelsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	lr.HandleFunc("/admin/users/levels", linkerAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Allow"))
	lr.HandleFunc("/admin/users/levels", linkerAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Disallow"))

	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarksPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/mine", bookmarksMinePage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Save"))
	bmr.HandleFunc("/edit", bookmarksEditCreateActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Create"))
	bmr.HandleFunc("/edit", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiresAnAccount())

	ibr := r.PathPrefix("/imagebbs").Subrouter()
	//ibr.HandleFunc(".rss", imagebbsRssPage).Methods("GET")
	//ibr.HandleFunc("/board/{boardno:[0-9+}.rss", imagebbsBoardRssPage).Methods("GET")
	//ibr.HandleFunc(".atom", imagebbsAtomPage).Methods("GET")
	//ibr.HandleFunc("/board/{boardno:[0-9+}.atom", imagebbsBoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbsBoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbsBoardPostImageActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Add offsite image"))
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbsBoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbsBoardThreadReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Reply"))
	ibr.HandleFunc("", imagebbsPage).Methods("GET")
	ibr.HandleFunc("/admin", imagebbsAdminPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", imagebbsAdminBoardsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/boards", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbsAdminNewBoardPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board", imagebbsAdminNewBoardMakePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("New board"))
	ibr.HandleFunc("/admin/board", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
	ibr.HandleFunc("/admin/board/{board}", imagebbsAdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Modify board"))

	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", searchPage).Methods("GET")
	sr.HandleFunc("", searchResultForumActionPage).Methods("POST").MatcherFunc(TaskMatcher("Search forum"))
	sr.HandleFunc("", searchResultNewsActionPage).Methods("POST").MatcherFunc(TaskMatcher("Search news"))
	sr.HandleFunc("", searchResultLinkerActionPage).Methods("POST").MatcherFunc(TaskMatcher("Search linker"))
	sr.HandleFunc("", searchResultBlogsActionPage).Methods("POST").MatcherFunc(TaskMatcher("Search blogs"))
	sr.HandleFunc("", searchResultWritingsActionPage).Methods("POST").MatcherFunc(TaskMatcher("Search writings"))

	wr := r.PathPrefix("/writings").Subrouter()
	//wr.HandleFunc(".rss", writingsRssPage).Methods("GET")
	//wr.HandleFunc(".atom", writingsAtomPage).Methods("GET")
	wr.HandleFunc("", writingsPage).Methods("GET")
	wr.HandleFunc("/", writingsAdminCategoriesModifyPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Modify Category"))
	wr.HandleFunc("/", writingsAdminCategoriesCreatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("New Category"))
	wr.HandleFunc("/article/{article}", writingsArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", writingsArticleReplyActionPage).Methods("POST").MatcherFunc(TaskMatcher("Reply"))
	wr.HandleFunc("/article/{article}/edit", writingsArticleEditPage).Methods("GET").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator")))
	wr.HandleFunc("/article/{article}/edit", writingsArticleEditActionPage).Methods("POST").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher("Update writing"))
	wr.HandleFunc("/categories", writingsCategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", writingsCategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", writingsCategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", writingsArticleAddPage).Methods("GET").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", writingsArticleAddActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher("Submit writing"))
	wr.HandleFunc("/admin/users/levels", writingsAdminUserLevelsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	wr.HandleFunc("/admin/users/levels", writingsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Allow"))
	wr.HandleFunc("/admin/users/levels", writingsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Disallow"))
	wr.HandleFunc("/admin/users/access", writingsAdminUserAccessPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	wr.HandleFunc("/admin/users/access", writingsAdminUserAccessAddActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Add approval"))
	wr.HandleFunc("/admin/users/access", writingsAdminUserAccessUpdateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update user approval"))
	wr.HandleFunc("/admin/users/access", writingsAdminUserAccessRemoveActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete user approval"))
	wr.HandleFunc("/admin/categories", writingsAdminCategoriesPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	wr.HandleFunc("/admin/categories", writingsAdminCategoriesModifyPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("writing category change"))
	wr.HandleFunc("/admin/categories", writingsAdminCategoriesCreatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("writing category create"))

	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", informationPage).Methods("GET")

	ur := r.PathPrefix("/user").Subrouter()
	ur.HandleFunc("", userPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/logout", userLogoutPage).Methods("GET")
	ur.HandleFunc("/lang", userLangPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/lang", userLangSaveLanguagesActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Save languages"))
	ur.HandleFunc("/email", userEmailPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/email", userEmailSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Save all"))
	ur.HandleFunc("/email", userEmailTestActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Test mail"))

	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", registerPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", registerActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(TaskMatcher("Register"))

	ulr := r.PathPrefix("/login").Subrouter()
	ulr.HandleFunc("", loginPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	ulr.HandleFunc("", loginActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(TaskMatcher("Login"))

	ar := r.PathPrefix("/admin").MatcherFunc(RequiredAccess("administrator")).Subrouter()
	ar.Use(AdminCheckerMiddleware)
	ar.HandleFunc("", adminPage).Methods("GET")
	ar.HandleFunc("/", adminPage).Methods("GET")
	ar.HandleFunc("/forum", adminForumPage).Methods("GET")
	ar.HandleFunc("/forum", adminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(TaskMatcher("Remake statistic information on forumthread"))
	ar.HandleFunc("/forum", adminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(TaskMatcher("Remake statistic information on forumtopic"))
	ar.HandleFunc("/forum/list", adminForumWordListPage).Methods("GET")
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users", adminUsersDoNothingPage).Methods("POST").MatcherFunc(TaskMatcher("User do nothing"))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsUserAllowPage).Methods("POST").MatcherFunc(TaskMatcher("User Allow"))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher("User Disallow"))
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/languages", adminLanguagesRenamePage).Methods("POST").MatcherFunc(TaskMatcher("Rename Language"))
	ar.HandleFunc("/languages", adminLanguagesDeletePage).Methods("POST").MatcherFunc(TaskMatcher("Delete Language"))
	ar.HandleFunc("/languages", adminLanguagesCreatePage).Methods("POST").MatcherFunc(TaskMatcher("Create Language"))
	ar.HandleFunc("/search", adminSearchPage).Methods("GET")
	ar.HandleFunc("/search", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(TaskMatcher("Remake comments search"))
	ar.HandleFunc("/search", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(TaskMatcher("Remake news search"))
	ar.HandleFunc("/search", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(TaskMatcher("Remake blog search"))
	ar.HandleFunc("/search", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(TaskMatcher("Remake linker search"))
	ar.HandleFunc("/search", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(TaskMatcher("Remake writing search"))
	ar.HandleFunc("/search/list", adminSearchWordListPage).Methods("GET")

	// oauth shit
	//r.HandleFunc("/", homePage)
	//r.HandleFunc("/login", loginPage)
	//r.HandleFunc("/callback", callbackHandler)
	//r.HandleFunc("/logout", logoutHandler)

	http.Handle("/", r)

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// TODO we could do better
func TargetUsersLevelNotHigherThanAdminsMax() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		session := r.Context().Value(ContextValues("session")).(*sessions.Session)
		adminUid, _ := session.Values["UID"].(int32)

		targetUid, err := strconv.Atoi(r.PostFormValue("uid"))
		if err != nil {
			return false
		}

		tid, err := strconv.Atoi(r.PostFormValue("tid"))
		if err != nil {
			return false
		}

		queries := r.Context().Value(ContextValues("queries")).(*Queries)

		targetUser, err := queries.getUsersTopicLevel(r.Context(), getUsersTopicLevelParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(targetUid),
		})
		if err != nil {
			return false
		}

		adminUser, err := queries.getUsersTopicLevel(r.Context(), getUsersTopicLevelParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(adminUid),
		})
		if err != nil {
			return false
		}

		return adminUser.Invitemax.Int32 >= targetUser.Level.Int32
	}
}

// TODO we could do better
func AdminUsersMaxLevelNotLowerThanTargetLevel() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		session := r.Context().Value(ContextValues("session")).(*sessions.Session)
		adminUid, _ := session.Values["UID"].(int32)

		inviteMax, err := strconv.Atoi(r.PostFormValue("inviteMax"))
		if err != nil {
			return false
		}
		level, err := strconv.Atoi(r.PostFormValue("level"))
		if err != nil {
			return false
		}
		tid, err := strconv.Atoi(r.PostFormValue("tid"))
		if err != nil {
			return false
		}
		queries := r.Context().Value(ContextValues("queries")).(*Queries)

		adminUser, err := queries.getUsersTopicLevel(r.Context(), getUsersTopicLevelParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(adminUid),
		})
		if err != nil {
			return false
		}

		return int(adminUser.Invitemax.Int32) >= level && int(adminUser.Invitemax.Int32) >= inviteMax
	}
}

func RequiredAccess(accessLevels ...string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		// TODO
		return true
	}
}

func RequiresAnAccount() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		// TODO
		return true
	}
}

func NewsPostAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		newsPostId, _ := strconv.Atoi(vars["post"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session := request.Context().Value(ContextValues("session")).(*sessions.Session)
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.getNewsThreadId(request.Context(), int32(newsPostId))
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.Idusers.Int32 == uid
	}
}

func BlogAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		blogId, _ := strconv.Atoi(vars["blog"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session := request.Context().Value(ContextValues("session")).(*sessions.Session)
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.show_blog(request.Context(), int32(blogId))
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}

func WritingAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		writingId, _ := strconv.Atoi(vars["writing"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session := request.Context().Value(ContextValues("session")).(*sessions.Session)
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.fetchWritingById(request.Context(), fetchWritingByIdParams{
			Userid:    uid,
			Idwriting: int32(writingId),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}

func CommentAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		commentId, _ := strconv.Atoi(vars["comment"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session := request.Context().Value(ContextValues("session")).(*sessions.Session)
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.user_get_comment(request.Context(), user_get_commentParams{
			UsersIdusers: uid,
			Idcomments:   int32(commentId),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}

func TaskMatcher(taskName string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return request.PostFormValue("task") == taskName
	}
}

func NoTask() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return request.PostFormValue("task") == ""
	}
}

//func oauthHomeHandler(w http.ResponseWriter, r *http.Request) {
//	// Check if user is authenticated
//	if !isAuthenticated(r) {
//		http.Redirect(w, r, "/login", http.StatusFound)
//		return
//	}
//
//	tmpl := `
//		<!DOCTYPE html>
//		<html>
//		<head>
//			<title>Home Page</title>
//		</head>
//		<body>
//			<h1>Welcome, {{ .Email }}</h1>
//			<a href="/logout">Logout</a>
//		</body>
//		</html>
//	`
//
//	t := template.Must(template.New("home").Parse(tmpl))
//	data := map[string]string{"Email": getEmail(r)}
//
//	t.Execute(w, data)
//}
//
//func loginHandler(w http.ResponseWriter, r *http.Request) {
//	// Generate the URL to redirect the user to Google's consent page
//	url := oauth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
//	http.Redirect(w, r, url, http.StatusFound)
//}
//
//func callbackHandler(w http.ResponseWriter, r *http.Request) {
//	code := r.FormValue("code")
//	if code == "" {
//		http.Error(w, "Failed to get authorization code", http.StatusInternalServerError)
//		return
//	}
//
//	// Exchange the authorization code for an access token
//	token, err := oauth2Config.Exchange(context.Background(), code)
//	if err != nil {
//		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
//		return
//	}
//
//	// Store the access token in the session
//	session, _ := store.Get(r, sessionName)
//	session.Values[sessionKey] = token.AccessToken
//	session.Save(r, w)
//
//	http.Redirect(w, r, "/", http.StatusFound)
//}
//
//func logoutHandler(w http.ResponseWriter, r *http.Request) {
//	// Clear the session and log the user out
//	session, _ := store.Get(r, sessionName)
//	session.Values[sessionKey] = nil
//	session.Save(r, w)
//
//	http.Redirect(w, r, "/", http.StatusFound)
//}
//
//func isAuthenticated(r *http.Request) bool {
//	session, _ := store.Get(r, sessionName)
//	accessToken, ok := session.Values[sessionKey]
//	if !ok {
//		return false
//	}
//
//	return accessToken != nil
//}
//
//func getEmail(r *http.Request) string {
//	// Fetch user's email using the access token from the session
//	session, _ := store.Get(r, sessionName)
//	_, ok := session.Values[sessionKey]
//	if !ok {
//		return ""
//	}
//
//	// Here, you can use the access token to fetch the user's email from the Google API
//	// For simplicity, we just return a dummy email
//	return "example@example.com"
//}
