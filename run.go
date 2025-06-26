package goa4web

import (
	"context"
	"fmt"
	common "github.com/arran4/goa4web/core/common"
	corelanguage "github.com/arran4/goa4web/core/language"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	hcommon "github.com/arran4/goa4web/handlers/common"
	news "github.com/arran4/goa4web/handlers/news"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	email "github.com/arran4/goa4web/internal/email"
	middleware "github.com/arran4/goa4web/internal/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/pkg/server"
	"github.com/arran4/goa4web/runtimeconfig"
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
	srv         *server.Server

	version = "dev"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// RunWithConfig starts the application using the provided configuration and
// session secret. The context controls the lifetime of the HTTP server.
func RunWithConfig(ctx context.Context, cfg runtimeconfig.RuntimeConfig, sessionSecret string) error {
	store = sessions.NewCookieStore([]byte(sessionSecret))
	core.Store = store
	core.SessionName = sessionName
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	common.Version = version

	var handler http.Handler

	if err := performStartupChecks(cfg); err != nil {
		return fmt.Errorf("startup checks: %w", err)
	}

	if err := corelanguage.ValidateDefaultLanguage(context.Background(), New(dbPool), cfg.DefaultLanguage); err != nil {
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
		userhandlers.UserAdderMiddleware,
		CoreAdderMiddleware,
		RequestLoggerMiddleware,
		middleware.SecurityHeadersMiddleware,
	).Wrap(r)
	if csrfEnabled() {
		handler = newCSRFMiddleware(sessionSecret, cfg.HTTPHostname, version).Wrap(handler)
	}

	srv = server.New(handler, store, dbPool, cfg)
	adminhandlers.ConfigFile = ConfigFile
	adminhandlers.Srv = srv
	adminhandlers.DBPool = dbPool
	adminhandlers.UpdateConfigKeyFunc = UpdateConfigKey

	provider := email.ProviderFromConfig(cfg)

	startWorkers(ctx, dbPool, provider)

	if err := server.Run(ctx, srv, cfg.HTTPListen); err != nil {
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
			CoreData: r.Context().Value(hcommon.KeyCoreData).(*CoreData),
		}

		news.CustomNewsIndex(data.CoreData, r)

		log.Printf("rendering template %s", template)

		if err := templates.RenderTemplate(w, template, data, common.NewFuncs(r)); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}

func AddNewsIndex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(hcommon.KeyCoreData).(*CoreData)
		news.CustomNewsIndex(cd, r)
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

// TODO we could do better

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
