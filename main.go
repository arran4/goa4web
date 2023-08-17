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

	r.HandleFunc("/main.css", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(getMainCSSData())
	}).Methods("GET")

	// News
	r.HandleFunc("/", newsPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	//nr.HandleFunc(".rss", newsRssPage).Methods("GET")
	nr.HandleFunc("", newsPage).Methods("GET")
	//nr.HandleFunc("{id:[0-9]+}", newsPostPage).Methods("GET")

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

	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("", forumPage).Methods("GET")
	fr.HandleFunc("/category/{category}", forumPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}", forumTopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread/{thread}", forumThreadPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread/{thread}", taskDoneAutoRefreshPage).Methods("POST")
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyPage).Methods("POST")
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumPage).Methods("POST").MatcherFunc(TaskMatcher("Reply"))
	br.HandleFunc("/topic/{topic}/thread/{thread}/reply", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher("Cancel"))
	/*TODO*/ fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}/edit", forumTopicThreadCommentEditPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}/edit", taskDoneAutoRefreshPage).Methods("POST")
	/*TODO*/ fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}/edit", forumTopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(TaskMatcher("Edit Post"))
	/*TODO*/ fr.HandleFunc("/topic/{topic}/new", forumPage).Methods("GET")
	/*TODO*/ fr.HandleFunc("/topic/{topic}/new", forumPage).Methods("POST").MatcherFunc(TaskMatcher("Create"))
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
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Set user level"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update user level"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete user level"))
	fr.HandleFunc("/admin/user/{user}/levels", forumAdminUserLevelPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete user level"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update user level"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Set user level"))
	fr.HandleFunc("/admin/restrictions/users", forumAdminUsersRestrictionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Update topic restriction"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Delete topic restriction"))
	fr.HandleFunc("/admin/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("Set topic restriction"))

	lr := r.PathPrefix("/linker").Subrouter()
	//lr.HandleFunc(".rss", linkerRssPage).Methods("GET")
	//lr.HandleFunc(".atom", linkerAtomPage).Methods("GET")
	lr.HandleFunc("", linkerPage).Methods("GET")

	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarksPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/mine", bookmarksMinePage).Methods("GET", "POST").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Save"))
	bmr.HandleFunc("/edit", bookmarksEditCreateActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher("Create"))

	ibr := r.PathPrefix("/imagebbs").Subrouter()
	//ibr.HandleFunc(".rss", imagebbsRssPage).Methods("GET")
	//ibr.HandleFunc("/board/{boardno:[0-9+}.rss", imagebbsBoardRssPage).Methods("GET")
	//ibr.HandleFunc(".atom", imagebbsAtomPage).Methods("GET")
	//ibr.HandleFunc("/board/{boardno:[0-9+}.atom", imagebbsBoardAtomPage).Methods("GET")
	ibr.HandleFunc("", imagebbsPage).Methods("GET")

	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", searchPage).Methods("GET")

	wr := r.PathPrefix("/writings").Subrouter()
	//wr.HandleFunc(".rss", writingsRssPage).Methods("GET")
	//wr.HandleFunc(".atom", writingsAtomPage).Methods("GET")
	wr.HandleFunc("", writingsPage).Methods("GET")

	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", informationPage).Methods("GET")

	ur := r.PathPrefix("/user").Subrouter()
	ur.HandleFunc("", userPage).Methods("GET")

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
