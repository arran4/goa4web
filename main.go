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
	}
}

func main() {
	r := mux.NewRouter()

	r.Use(DBAdderMiddleware)
	r.Use(UserAdderMiddleware)
	r.Use(CoreAdderMiddleware)

	r.HandleFunc("/main.css", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(mainCSSData)
	}).Methods("GET")

	r.HandleFunc("/", newsHandler).Methods("GET")
	nr := r.PathPrefix("/news").Subrouter()
	nr.HandleFunc("", newsHandler).Methods("GET")
	//nr.HandleFunc("{id:[0-9]+}", newsPostHandler).Methods("GET")

	ar := r.PathPrefix("/admin").Subrouter()
	ar.Use(AdminCheckerMiddleware)
	ar.HandleFunc("", adminHandler).Methods("GET")
	ar.HandleFunc("/", adminHandler).Methods("GET")
	ar.HandleFunc("/forum", adminForumHandler).Methods("GET")
	ar.HandleFunc("/forum", adminForumRemakeForumThreadHandler).Methods("POST").MatcherFunc(taskMatcher("Remake statistic information on forumthread"))
	ar.HandleFunc("/forum", adminForumRemakeForumTopicHandler).Methods("POST").MatcherFunc(taskMatcher("Remake statistic information on forumtopic"))
	ar.HandleFunc("/forum/list", adminForumWordListHandler).Methods("GET")
	//ar.HandleFunc("/users", adminUserHandler).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUserPermissionsHandler).Methods("GET")
	ar.HandleFunc("/languages", adminLanguageHandler).Methods("GET")
	ar.HandleFunc("/languages", adminLanguageRenameHandler).Methods("POST").MatcherFunc(taskMatcher("Rename Language"))
	ar.HandleFunc("/languages", adminLanguageDeleteHandler).Methods("POST").MatcherFunc(taskMatcher("Delete Language"))
	ar.HandleFunc("/languages", adminLanguageCreateHandler).Methods("POST").MatcherFunc(taskMatcher("Create Language"))
	//ar.HandleFunc("/search", adminSearchHandler).Methods("GET")
	//ar.HandleFunc("/forum", adminForumHandler).Methods("GET")

	// oauth shit
	//r.HandleFunc("/", homeHandler)
	//r.HandleFunc("/login", loginHandler)
	//r.HandleFunc("/callback", callbackHandler)
	//r.HandleFunc("/logout", logoutHandler)

	http.Handle("/", r)

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
