package goa4web

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConfigFile stores the path to the configuration file if provided on the
// command line. It is used by admin handlers when updating settings.
var ConfigFile string

var (
	//      // Replace these with your Google OAuth2 credentials
	//      clientID     = ""
	//      clientSecret = ""
	//      redirectURL  = "http://localhost:8080/callback"
	//
	//      // Change this to your desired session key
	sessionName = "my-session"
	store       *sessions.CookieStore
	srv         *Server

	version = "dev"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// RunWithConfig starts the application using the provided configuration and
// session secret. The context controls the lifetime of the HTTP server.
func RunWithConfig(ctx context.Context, cfg RuntimeConfig, sessionSecret string) error {
	store = sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	var handler http.Handler

	if err := performStartupChecks(cfg); err != nil {
		return fmt.Errorf("startup checks: %w", err)
	}

	if err := validateDefaultLanguage(context.Background(), New(dbPool), cfg.DefaultLanguage); err != nil {
		return fmt.Errorf("default language: %w", err)
	}

	if dbPool != nil {
		defer func() {
			if err := dbPool.Close(); err != nil {
				log.Printf("DB close error: %v", err)
			}
		}()
	}

	r := mux.NewRouter()
	registerRoutes(r)

	handler = newMiddlewareChain(
		DBAdderMiddleware,
		UserAdderMiddleware,
		CoreAdderMiddleware,
		RequestLoggerMiddleware,
		SecurityHeadersMiddleware,
	).Wrap(r)
	if csrfEnabled() {
		handler = newCSRFMiddleware(sessionSecret, cfg.HTTPHostname, version).Wrap(handler)
	}

	srv = newServer(handler, store, dbPool, cfg)

	provider := providerFromConfig(cfg)

	startWorkers(ctx, dbPool, provider)

	if err := runServer(ctx, srv, cfg.HTTPListen); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func runTemplate(template string) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type Data struct {
			*CoreData
		}

		data := Data{
			CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		}

		CustomNewsIndex(data.CoreData, r)

		log.Printf("rendering template %s", template)

		if err := renderTemplate(w, r, template, data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}

func AddNewsIndex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
		CustomNewsIndex(cd, r)
		handler.ServeHTTP(w, r)
	})
}

// safeGo runs fn in a goroutine and terminates the program if a panic occurs.
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panic: %v", r)
				os.Exit(1)
			}
		}()
		fn()
	}()
}

// mainCSSHandler serves the site's stylesheet.
func mainCSSHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "main.css", time.Time{}, bytes.NewReader(getMainCSSData()))
}

// redirectPermanent returns a handler that redirects to the provided path using
// StatusPermanentRedirect to preserve the request method.
func redirectPermanent(to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusPermanentRedirect)
	}
}

// redirectPermanentPrefix redirects any path starting with the given prefix to
// the same path under a new prefix while preserving query parameters.
func redirectPermanentPrefix(from, to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, from)
		if !strings.HasPrefix(rest, "/") && rest != "" {
			rest = "/" + rest
		}
		target := to + rest
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
	}
}

// TODO we could do better
func TargetUsersLevelNotHigherThanAdminsMax() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		session, err := GetSession(r)
		if err != nil {
			return false
		}
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

		targetUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(targetUid),
		})
		if err != nil {
			return false
		}

		adminUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), GetUsersTopicLevelByUserIdAndThreadIdParams{
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
		session, err := GetSession(r)
		if err != nil {
			return false
		}
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

		adminUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), GetUsersTopicLevelByUserIdAndThreadIdParams{
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
		return roleAllowed(request, accessLevels...)
	}
}

func RequiresAnAccount() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)
		return uid != 0
	}
}

func NewsPostAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		newsPostId, _ := strconv.Atoi(vars["post"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetForumThreadIdByNewsPostId(request.Context(), int32(newsPostId))
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
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetBlogEntryForUserById(request.Context(), int32(blogId))
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: %s", err)
				return false
			}
		}

		return row.UsersIdusers == uid
	}
}

func WritingAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		writingId, _ := strconv.Atoi(vars["writing"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetWritingByIdForUserDescendingByPublishedDate(request.Context(), GetWritingByIdForUserDescendingByPublishedDateParams{
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
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetCommentByIdForUser(request.Context(), GetCommentByIdForUserParams{
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

func GetThreadAndTopic() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		topicID, err := strconv.Atoi(vars["topic"])
		if err != nil {
			return false
		}
		threadID, err := strconv.Atoi(vars["thread"])
		if err != nil {
			return false
		}

		queries := r.Context().Value(ContextValues("queries")).(*Queries)

		session, _ := GetSession(r)
		var uid int32
		if session != nil {
			uid, _ = session.Values["UID"].(int32)
		}

		threadRow, err := queries.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissions(r.Context(), GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams{
			UsersIdusers:  uid,
			Idforumthread: int32(threadID),
		})
		if err != nil {
			log.Printf("GetThreadAndTopic thread: %v", err)
			return false
		}

		topicRow, err := queries.GetForumTopicByIdForUser(r.Context(), GetForumTopicByIdForUserParams{
			UsersIdusers: uid,
			Idforumtopic: threadRow.ForumtopicIdforumtopic,
		})
		if err != nil {
			log.Printf("GetThreadAndTopic topic: %v", err)
			return false
		}

		if int(topicRow.Idforumtopic) != topicID {
			return false
		}

		ctx := context.WithValue(r.Context(), ContextValues("thread"), threadRow)
		ctx = context.WithValue(ctx, ContextValues("topic"), topicRow)
		*r = *r.WithContext(ctx)
		return true
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
