package main

import (
	"embed"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
	//go:embed "templates/*.tmpl"
	templateFS        embed.FS
	compiledTemplates = template.Must(template.New("").Funcs(NewFuncs()).ParseFS(templateFS, "templates/*.tmpl"))
	//go:embed "main.css"
	mainCSSData []byte
)

func NewFuncs() template.FuncMap {
	return map[string]any{
		//"getSecurityLevel":
		"now": func() time.Time { return time.Now() },
		"a4code2html": func(s string) template.HTML {
			c := a4code2html{}
			c.input = s
			c.process()
			return template.HTML(c.output.String())
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
		_, _ = writer.Write(mainCSSData)
	}).Methods("GET")

	// News
	r.HandleFunc("/", newsPage).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	//nr.HandleFunc(".rss", newsRssPage).Methods("GET")
	nr.HandleFunc("", newsPage).Methods("GET")
	//nr.HandleFunc("{id:[0-9]+}", newsPostPage).Methods("GET")

	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.HandleFunc("", faqPage).Methods("GET")

	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", blogsRssPage).Methods("GET")
	br.HandleFunc("/atom", blogsAtomPage).Methods("GET")
	br.HandleFunc("", blogsPage).Methods("GET")
	br.HandleFunc("/user/permissions", blogsUserPermissionsPage).Methods("GET").MatcherFunc(requiredAccess("administrator"))
	br.HandleFunc("/add", blogsAddBlogPage).Methods("GET").MatcherFunc(requiredAccess("writer"))
	br.HandleFunc("/bloggers", blogsBloggersPage).Methods("GET")
	br.HandleFunc("/blogs/blogger/{blogger}", blogsBloggersPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", blogsBlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}/comments", blogsCommentPage).Methods("GET")
	br.HandleFunc("/blog/{blog}/comment/{comment}/edit", blogsCommentEditPage).Methods("GET")
	br.HandleFunc("/blog/{blog}/comment/{comment}/reply", blogsCommentReplyPage).Methods("GET")
	br.HandleFunc("/blog/{blog}/comment/{comment}/reply", blogsCommentReplyFullPage).Queries("type", "full").Methods("GET")
	br.HandleFunc("/blog/{blog}/edit", blogsEditBlogPage).Methods("GET").MatcherFunc(requiredAccess("writer"))

	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("", forumPage).Methods("GET")

	lr := r.PathPrefix("/linker").Subrouter()
	//lr.HandleFunc(".rss", linkerRssPage).Methods("GET")
	//lr.HandleFunc(".atom", linkerAtomPage).Methods("GET")
	lr.HandleFunc("", linkerPage).Methods("GET")

	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarksPage).Methods("GET")

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

	ar := r.PathPrefix("/admin").Subrouter()
	ar.Use(AdminCheckerMiddleware)
	ar.HandleFunc("", adminPage).Methods("GET")
	ar.HandleFunc("/", adminPage).Methods("GET")
	ar.HandleFunc("/forum", adminForumPage).Methods("GET")
	ar.HandleFunc("/forum", adminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(taskMatcher("Remake statistic information on forumthread"))
	ar.HandleFunc("/forum", adminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(taskMatcher("Remake statistic information on forumtopic"))
	ar.HandleFunc("/forum/list", adminForumWordListPage).Methods("GET")
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users", adminUsersDoNothingPage).Methods("POST").MatcherFunc(taskMatcher("User do nothing"))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsUserAllowPage).Methods("POST").MatcherFunc(taskMatcher("User Allow"))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(taskMatcher("User Disallow"))
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/languages", adminLanguagesRenamePage).Methods("POST").MatcherFunc(taskMatcher("Rename Language"))
	ar.HandleFunc("/languages", adminLanguagesDeletePage).Methods("POST").MatcherFunc(taskMatcher("Delete Language"))
	ar.HandleFunc("/languages", adminLanguagesCreatePage).Methods("POST").MatcherFunc(taskMatcher("Create Language"))
	ar.HandleFunc("/search", adminSearchPage).Methods("GET")
	ar.HandleFunc("/search", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(taskMatcher("Remake comments search"))
	ar.HandleFunc("/search", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(taskMatcher("Remake news search"))
	ar.HandleFunc("/search", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(taskMatcher("Remake blog search"))
	ar.HandleFunc("/search", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(taskMatcher("Remake linker search"))
	ar.HandleFunc("/search", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(taskMatcher("Remake writing search"))
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

func requiredAccess(accessLevels ...string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		// TODO
		return true
	}
}

func taskMatcher(taskName string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return request.PostFormValue("task") == taskName
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
