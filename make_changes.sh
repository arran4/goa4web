# 1. core/session.go
cat << 'CORE_EOF' > core/session.go
package core

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
)

var SessionName string
var Store *sessions.CookieStore

func GetSession(r *http.Request) (*sessions.Session, error) {
	if sessVal := r.Context().Value(ContextValues("session")); sessVal != nil {
		sess, ok := sessVal.(*sessions.Session)
		if !ok {
			return nil, fmt.Errorf("invalid session in context")
		}
		return sess, nil
	}
	sess, err := Store.Get(r, SessionName)
	if err != nil {
		log.Printf("get session: %v", err)
	}
	return sess, err
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	sess, _ := Store.New(r, SessionName)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		log.Printf("clear session: %v", err)
	}
}

func GetSessionOrFail(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	sess, err := GetSession(r)
	if err != nil {
		SessionErrorRedirect(w, r, err)
		return nil, false
	}
	return sess, true
}

func SessionErrorRedirect(w http.ResponseWriter, r *http.Request, err error) {
	SessionError(w, r, err)
	RedirectToLogin(w, r, nil)
}

func SessionError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("session error: %v", err)
	clearSession(w, r)
}

func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Cloudflare-CDN-Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	DisableCaching(w)
	if session != nil {
		if err := session.Save(r, w); err != nil {
			log.Printf("save session: %v", err)
		}
	}

	vals := r.URL.Query()
	back := vals.Get("back")
	if back == "" {
		back = r.URL.RequestURI()
	}

	newVals := url.Values{}
	newVals.Set("back", back)
	if r.Method != http.MethodGet {
		newVals.Set("method", r.Method)
	}

	http.Redirect(w, r, "/login?"+newVals.Encode(), http.StatusSeeOther)
}

type ContextValues string
CORE_EOF

# 2. handlers/auth/login_task.go
cat << 'LT_EOF' > handlers/auth/login_task.go
package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type LoginTask struct {
	tasks.TaskString
}

var loginTask = &LoginTask{TaskString: TaskLogin}

var _ tasks.Task = (*LoginTask)(nil)
var _ tasks.TemplatesRequired = (*LoginTask)(nil)

const (
	templateLoginPage          = "pages/auth/loginPage.gohtml"
	templatePasswordVerifyPage = "pages/auth/passwordVerifyPage.gohtml"
)

func (LoginTask) Page(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, r.URL.Query().Get("error"), r.URL.Query().Get("notice"))
}

func (LoginTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		sess, _ := core.GetSession(r)
		log.Printf("login attempt for %s session=%s", r.PostFormValue("username"), handlers.HashSessionID(sess.ID))
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	queries := cd.Queries()

	cfg := cd.Config
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if cfg.LoginAttemptThreshold > 0 {
		since := time.Now().Add(-time.Duration(cfg.LoginAttemptWindow) * time.Minute)
		cnt, err := queries.SystemCountRecentLoginAttempts(r.Context(), db.SystemCountRecentLoginAttemptsParams{Username: username, IpAddress: ip, CreatedAt: since})
		if err != nil {
			log.Printf("count login attempts: %v", err)
		} else if cnt >= int64(cfg.LoginAttemptThreshold) {
			return loginFormHandler{msg: "Too many failed attempts"}
		}
	}

	session := cd.GetSession()
	alreadyLoggedInMsg := ""
	if _, ok := session.Values["UID"].(int32); ok {
		alreadyLoggedInMsg = " You remain logged in as your current account."
	}

	row, err := cd.UserCredentials(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := queries.SystemInsertLoginAttempt(r.Context(), db.SystemInsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
				log.Printf("insert login attempt: %v", err)
			}
			return loginFormHandler{msg: "Invalid username or password." + alreadyLoggedInMsg}
		}
		return fmt.Errorf("LoginTask.Action: user credentials query: %w", err)
	}

	if !VerifyPassword(password, row.Passwd.String, row.PasswdAlgorithm.String) {
		expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
		reset, err := queries.GetPasswordResetByUser(r.Context(), db.GetPasswordResetByUserParams{UserID: row.Idusers, CreatedAt: expiry})
		if err == nil && VerifyPassword(password, reset.Passwd, reset.PasswdAlgorithm) {
			code := r.FormValue("code")
			if code != "" {
				if err := cd.VerifyPasswordReset(code, password); err != nil {
					if err := queries.SystemInsertLoginAttempt(r.Context(), db.SystemInsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
						log.Printf("insert login attempt: %v", err)
					}
					return loginFormHandler{msg: "Invalid username or password." + alreadyLoggedInMsg}
				}
			} else {
				type Data struct {
					ID int32
				}
				cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
				cd.PageTitle = "Verify Password"
				data := Data{ID: reset.ID}
				return handlers.TemplateWithDataHandler(templatePasswordVerifyPage, data)
			}
		} else {
			if err := queries.SystemInsertLoginAttempt(r.Context(), db.SystemInsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
				log.Printf("insert login attempt: %v", err)
			}
			return loginFormHandler{msg: "Invalid username or password." + alreadyLoggedInMsg}
		}
	}

	if _, err := queries.GetLoginRoleForUser(r.Context(), row.Idusers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return loginFormHandler{msg: "Approval is pending." + alreadyLoggedInMsg}
		}
		return fmt.Errorf("user role %w", err)
	}

	if row.PasswdAlgorithm.String == "" || row.PasswdAlgorithm.String == "md5" {
		newHash, newAlg, err := HashPassword(password)
		if err == nil {
			if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: row.Idusers, Passwd: newHash, PasswdAlgorithm: sql.NullString{String: newAlg, Valid: true}}); err != nil {
				log.Printf("insert password: %v", err)
			}
		}
	}

	// Fully authenticated. Now replace session A with session B.
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")

	if session.ID != "" && cd.SessionManager() != nil {
		_ = cd.SessionManager().DeleteSessionByID(r.Context(), session.ID)
	}

	session.Values["UID"] = int32(row.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	backURL, _ := cd.SanitizeBackURL(r, r.FormValue("back"))

	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("session save %w", err)
	}

	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("login success uid=%d session=%s", row.Idusers, handlers.HashSessionID(session.ID))
	}

	target := backURL
	if target == "" {
		target = "/"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusSeeOther)
	})
}

func (LoginTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{
		tasks.Template(templateLoginPage),
		tasks.Template(templatePasswordVerifyPage),
	}
}
LT_EOF

# 3. handlers/auth/loginPage.go
cat << 'LP_EOF' > handlers/auth/loginPage.go
package auth

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

type loginFormHandler struct{ msg string }

func (l loginFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, l.msg, "")
}

var _ http.Handler = (*loginFormHandler)(nil)

const (
	TaskDoneAutoRefreshPageTmpl tasks.Template = "pages/misc/taskDoneAutoRefreshPage.gohtml"
)

func renderLoginForm(w http.ResponseWriter, r *http.Request, errMsg, noticeMsg string) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.SetCurrentError(errMsg)
	cd.SetCurrentNotice(noticeMsg)
	type Data struct {
		Code    string
		Back    string
		Method  string
	}
	handlers.SetPageTitle(r, "Login")
	backURL, _ := cd.SanitizeBackURL(r, r.FormValue("back"))
	data := Data{
		Code:    r.FormValue("code"),
		Back:    backURL,
		Method:  r.FormValue("method"),
	}
	_ = LoginPageTmpl.Handle(w, r, data)
}

const LoginPageTmpl tasks.Template = "pages/auth/loginPage.gohtml"
LP_EOF

# 4. handlers/auth/routes.go
cat << 'AR_EOF' > handlers/auth/routes.go
package auth

import (
	"github.com/arran4/goa4web/handlers"
	gml "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

// RegisterRoutes attaches the login and registration endpoints to r.
func RegisterRoutes(r *mux.Router, _ *config.RuntimeConfig) []nav.RouterOptions {
	rr := r.PathPrefix("/register").Subrouter()
	rr.Use(handlers.IndexMiddleware(CustomIndex))
	rr.HandleFunc("", handlers.WithNoCache(registerTask.Page)).Methods("GET").MatcherFunc(gml.Not(handlers.RequiresAnAccount()))
	rr.HandleFunc("", handlers.TaskHandler(registerTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(registerTask.Matcher())

	lr := r.PathPrefix("/login").Subrouter()
	lr.HandleFunc("", handlers.WithNoCache(loginTask.Page)).Methods("GET")
	lr.HandleFunc("", handlers.TaskHandler(loginTask)).Methods("POST").MatcherFunc(loginTask.Matcher())
	lr.HandleFunc("/verify", handlers.TaskHandler(verifyPasswordTask)).Methods("POST").MatcherFunc(verifyPasswordTask.Matcher())

	lr.HandleFunc("/passkey/begin", HasWebAuthn(handlers.WithNoCache(loginPasskeyBegin))).Methods("GET")
	lr.HandleFunc("/passkey/finish", HasWebAuthn(loginPasskeyFinish)).Methods("POST")

	fr := r.PathPrefix("/forgot").Subrouter()
	fr.HandleFunc("", handlers.WithNoCache(forgotPasswordTask.Page)).Methods("GET").MatcherFunc(gml.Not(handlers.RequiresAnAccount()))
	fr.HandleFunc("", handlers.TaskHandler(emailAssociationRequestTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(emailAssociationRequestTask.Matcher())
	fr.HandleFunc("", handlers.TaskHandler(forgotPasswordTask)).Methods("POST").MatcherFunc(gml.Not(handlers.RequiresAnAccount())).MatcherFunc(forgotPasswordTask.Matcher())
	return nil
}

// Register registers the auth router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("auth", nil, func(r *mux.Router, cfg *config.RuntimeConfig) []nav.RouterOptions {
		return RegisterRoutes(r, cfg)
	})
}
AR_EOF

# 5. handlers/taskhandler.go
cat << 'TH_EOF' > handlers/taskhandler.go
package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

// TaskHandler wraps t.Action to record the task on the request event and handle the
// returned result
func TaskHandler(t tasks.Task) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(consts.KeyCoreData).(*common.CoreData); v != nil {
			v.SetEventTask(t)
		}
		result := t.Action(w, r)
		switch result := result.(type) {
		case RedirectHandler:
			// Use 303 See Other so POST actions redirect to a GET of the target resource.
			// 307 would preserve the HTTP method and often breaks when the target only supports GET.
			status := http.StatusSeeOther
			if r.Method == http.MethodGet {
				status = http.StatusTemporaryRedirect
			}
			http.Redirect(w, r, string(result), status)
		case RefreshDirectHandler:
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			cd.AutoRefresh = result.Content()
			_ = TaskDoneAutoRefreshPageTmpl.Handle(w, r, result)
		case TextByteWriter:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if _, err := w.Write([]byte(result)); err != nil {
				log.Printf("write response: %v", err)
			}
		case http.HandlerFunc:
			result(w, r)
		case http.Handler:
			result.ServeHTTP(w, r)
		case SessionFetchFail:
			loginRedirect(w, r)
		case *SessionFetchFail:
			loginRedirect(w, r)
		case nil:
			TaskDoneAutoRefreshPage(w, r)
		case error:
			var ue interface {
				error
				UserErrorMessage() string
			}
			if errors.As(result, &ue) {
				log.Printf("task action: %v", result)
				if msg := ue.UserErrorMessage(); msg != "" {
					r.URL.RawQuery = "error=" + url.QueryEscape(msg)
				} else {
					r.URL.RawQuery = "error=" + url.QueryEscape(result.Error())
				}
				TaskErrorAcknowledgementPage(w, r)
				return
			}
			log.Printf("task action: %v", result)
			RenderErrorPage(w, r, result)
			return
		default:
			RenderErrorPage(w, r, fmt.Errorf("%s", http.StatusText(http.StatusInternalServerError)))
		}
	}
}

func loginRedirect(w http.ResponseWriter, r *http.Request) {
	core.RedirectToLogin(w, r, nil)
}
TH_EOF

# 6. handlers/template.go
cat << 'HT_EOF' > handlers/template.go
package handlers

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// TemplateHandler renders the template and handles any template error.
// Example usage:
//
//	type Data struct{ Message string }
//	TemplateHandler(w, r, "page.gohtml", Data{"hello"})
//
// Template helpers are provided via the CoreData in the request context,
// accessible in templates as Funcs["cd"] (*common.CoreData).
func TemplateHandler(w http.ResponseWriter, r *http.Request, tmpl Page, data any) error {
	cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	sessionCookieName := "goa4web_session"
	if cd != nil && cd.Config != nil && cd.Config.SessionName != "" {
		sessionCookieName = cd.Config.SessionName
	}

	_, err := r.Cookie(sessionCookieName)
	hasCookie := err == nil

	if (cd != nil && cd.UserID != 0) || hasCookie {
		DisableCaching(w)
	}

	if err := tmpl.TemplateExecute(w, r, data); err != nil {
		log.Printf("Template Error: %s", err)
		errData := struct {
			Error   string
			BackURL string
		}{
			Error:   err.Error(),
			BackURL: r.Referer(),
		}
		if err2 := TaskErrorAcknowledgementPageTmpl.TemplateExecute(w, r, errData); err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			RenderErrorPage(w, r, common.ErrInternalServerError)
		}
		return err
	}
	return nil
}

// IndexMiddleware injects custom index items via fn before executing the next handler.
func IndexMiddleware(fn func(*common.CoreData, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
				fn(cd, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}
HT_EOF

# 7. handlers/user/userLogoutPage.go
cat << 'UL_EOF' > handlers/user/userLogoutPage.go
package user

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/core"
)

func userLogoutPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Logout"
	session, err := core.GetSession(r)
	if err != nil {
		core.SessionError(w, r, err)
	}
	uid, _ := session.Values["UID"].(int32)
	log.Printf("logout request session=%s uid=%d", handlers.HashSessionID(session.ID), uid)

	// session retrieved earlier
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")
	sm := cd.SessionManager()
	if session.ID != "" {
		if err := sm.DeleteSessionByID(r.Context(), session.ID); err != nil {
			log.Printf("delete session: %v", err)
		}
	}

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	log.Printf("logout success session=%s", handlers.HashSessionID(session.ID))

	clearLoggedOutCoreData(cd)

	handlers.DisableCaching(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearLoggedOutCoreData(cd *common.CoreData) {
	cd.UserID = 0
	cd.CustomIndexItems = nil
}
UL_EOF

# 8. internal/middleware/middleware.go
cat << 'MM_EOF' > internal/middleware/middleware.go
package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/sessions"
)

// RequestLoggerMiddleware logs incoming requests along with the user and session IDs.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := int32(0)
		sessID := ""
		var cd *common.CoreData
		if c, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && c != nil {
			cd = c
			uid = cd.UserID
			if s := cd.Session(); s != nil {
				sessID = s.ID
			}
		}
		if cd != nil && cd.Config != nil && cd.Config.LogFlags&config.LogFlagDebug != 0 {
			if r.URL.Path != "/ws/notifications" || uid != 0 {
				if sessID != "" {
					log.Printf("%s %s uid=%d session=%s", r.Method, r.URL.Path, uid, sessID)
				} else {
					log.Printf("%s %s uid=%d", r.Method, r.URL.Path, uid)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RecoverMiddleware logs panics from handlers and returns HTTP 500.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if goa4web.Version == "dev" {
				return
			}
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				handlers.RenderErrorPage(w, r, fmt.Errorf("%v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RedirectToLogin stores the current URL then redirects to the login page.
// It returns the HTTP status code used for the redirect.
func RedirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) int {
	core.RedirectToLogin(w, r, session)
	return http.StatusSeeOther
}
MM_EOF

# 9. core/templates/site/layouts/head.gohtml
cat << 'HEAD_EOF' > core/templates/site/layouts/head.gohtml
{{define "head"}}
{{ if cd.Marked "html-begin" }}
<!DOCTYPE html>
<html>
{{ end }}
    {{ if cd.Marked "head" }}
        <head>
       {{ if cd.PageTitle }}
       <title>{{cd.PageTitle}} - {{cd.SiteTitle}}</title>
       {{ else }}
       <title>{{cd.SiteTitle}}</title>
       {{ end }}
       {{template "headdata" cd}}
       <link rel="stylesheet" href="{{ assetHash "/main.css" }}">
       <link rel="icon" href="{{ assetHash "/favicon.svg" }}" type="image/svg+xml">
       <meta name="viewport" content="width=device-width, initial-scale=1">
       {{ if cd.HasModule "images" }}               <script src="{{ assetHash "/images/pasteimg.js" }}"></script>               {{ end }}
       {{ if and (cd.HasModule "websocket") cd.Config.NotificationsEnabled (ne cd.UserID 0) }}               <script src="{{ assetHash "/websocket/notifications.js" }}"></script>               {{ end }}
       <script src="{{ assetHash "/static/site.js" }}"></script>
       {{ if or cd.UserID (eq cd.PageTitle "Login") }}
       <script>
       window.addEventListener('pageshow', function(event) {
           if (event.persisted) {
               window.location.reload();
           }
       });
       </script>
       {{ end }}
       {{ if cd.AutoRefresh }}            <meta http-equiv="refresh" content="{{cd.AutoRefresh}}">       {{ end }}
       {{ if cd.UserID }}
       {{ if cd.CustomCSS }}
       <style>
       {{ cd.CustomCSS }}
       </style>
       {{ end }}
       {{ end }}
    </head>
    {{ end }}
    {{ if cd.Marked "bodyBegin" }}
    <body{{if cd.Section}} class="{{ cd.Section }}"{{end}}>
    {{ end }}
        {{ if cd.Marked "header" }}
            {{template "header"}}
        {{ end }}
    {{ if cd.Marked "body-table-begin" }}
        <input type="checkbox" id="nav-toggle" class="nav-toggle">
        <input type="checkbox" id="shrink-toggle" class="shrink-toggle">
        <div class="layout">
            <aside class="sidebar">{{template "index" $}}</aside>
            <main class="content">
    {{ end }}
        {{ if cd.Marked "announcements" }}
                    {{- with $a := cd.AnnouncementLoaded }}{{ if $a }}
                    <a href="/news/news/{{ $a.Idsitenews }}"><strong>{{ $a.News.String }}</strong></a><br />
                    {{- end }}{{ end }}
        {{ end }}

                    {{- end}}
HEAD_EOF

# 10. core/templates/site/pages/auth/loginPage.gohtml
cat << 'HTML_EOF' > core/templates/site/pages/auth/loginPage.gohtml
{{ template "head" $ }}
    {{- if cd.CurrentNotice }}
    <p>{{ cd.CurrentNotice }}</p>
    {{- end }}
    {{- if cd.CurrentError }}
    <p class="text-error">{{ cd.CurrentError }}</p>
    {{- end }}
    Login:<br>
    <form method="post" action="/login">
        {{ csrfField }}
        {{- if $.Back }}
        <input type="hidden" name="back" value="{{ $.Back }}">
        {{- end }}
        {{- if $.Method }}
        <input type="hidden" name="method" value="{{ $.Method }}">
        {{- end }}
        {{- if $.Code }}
        <input type="hidden" name="code" value="{{ $.Code }}">
        {{- end }}
        Username: <input name="username"><br>
        Password: <input name="password" type="password"><br>
        <input type="submit" name="task" value="Login">
        <br><a href="/forgot">Forgot password?</a>
    </form>
    <br>
    <hr>
    {{ if cd.WebAuthn }}<h3>Login with Passkey</h3>
    <form id="passkey-login-form">
        {{ csrfField }}
        Username: <input type="text" id="passkey-username" name="username" required><br>
        <button type="submit">Login with Passkey</button>
    </form>

    <script src="{{ assetHash "/static/passkeys.js" }}"></script>
{{ end }}
{{ template "tail" $ }}
HTML_EOF

# Tests updates

cat << 'T1_EOF' > core/session_test.go
package core_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core"
	"github.com/gorilla/sessions"
)

const sessionName = "test-session"

func TestGetSessionContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	session := &sessions.Session{Values: map[any]any{"foo": "bar"}}
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	req = req.WithContext(ctx)

	sess, err := core.GetSession(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
	if sess.Values["foo"] != "bar" {
		t.Errorf("expected 'bar', got %v", sess.Values["foo"])
	}
}

func TestGetSessionStore(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	sess, err := core.GetSession(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess == nil {
		t.Fatal("expected session, got nil")
	}
}

func TestSessionErrorRedirect(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	core.SessionErrorRedirect(rr, req, nil)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("expected cleared cookie, got %q", sc)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login?back=%2F" {
		t.Errorf("unexpected location %q", loc)
	}
}

func TestGetSessionOrFailBadSession(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "bad"})
	rr := httptest.NewRecorder()
	sess, ok := core.GetSessionOrFail(rr, req)
	if ok {
		t.Fatalf("expected failure, got session %v", sess)
	}
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	sc := rr.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("expected cleared cookie, got %q", sc)
	}
	loc := rr.Header().Get("Location")
	if loc != "/login?back=%2F" {
		t.Errorf("unexpected location %q", loc)
	}
}

func TestGetSessionOrFail(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = sessionName
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	sess, ok := core.GetSessionOrFail(rr, req)
	if !ok {
		t.Fatalf("expected success")
	}
	if sess == nil {
		t.Fatal("expected session")
	}
}
T1_EOF

cat << 'T2_EOF' > internal/middleware/middleware_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/sessions"
)

func TestRedirectToLogin(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	sess := testhelpers.Must(store.New(req, "sess"))
	rr := httptest.NewRecorder()
	code := RedirectToLogin(rr, req, sess)
	if code != http.StatusSeeOther {
		t.Fatalf("code=%d", code)
	}
	if rr.Result().StatusCode != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

func TestRedirectToLoginIncludesBackAndQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/page?x=1", nil)
	rr := httptest.NewRecorder()
	RedirectToLogin(rr, req, nil)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	q := u.Query()
	if got := q.Get("back"); got != "/page?x=1" {
		t.Fatalf("back=%q", got)
	}
	if q.Has("method") {
		t.Fatalf("unexpected method param: %s", q.Get("method"))
	}
	if q.Has("data") {
		t.Fatalf("unexpected data param: %s", q.Get("data"))
	}
}

func TestRedirectToLoginDiscardsPostData(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test"))
	form := url.Values{"a": {"1"}, "b": {"2"}}
	req := httptest.NewRequest(http.MethodPost, "/submit?foo=1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	sess := testhelpers.Must(store.New(req, "sess"))
	rr := httptest.NewRecorder()
	RedirectToLogin(rr, req, sess)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	q := u.Query()
	if got := q.Get("back"); got != "/submit?foo=1" {
		t.Fatalf("back=%q", got)
	}
	if got := q.Get("method"); got != http.MethodPost {
		t.Fatalf("method=%q", got)
	}
	if q.Has("data") {
		t.Fatalf("unexpected data parameter: %s", q.Get("data"))
	}
}
T2_EOF

cat << 'T3_EOF' > handlers/auth/issue3095_test.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"net/url"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

// Tests transition matrices for #3095.

func TestIssue3095_LoginRedirectOriginalDestination(t *testing.T) {
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "correcthorse")
	form.Set("back", "/protected-page")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         10,
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	q.GetLoginRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
		return 1, nil
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	handler, ok := result.(http.HandlerFunc)
	if !ok {
		t.Fatalf("Expected http.HandlerFunc from LoginTask.Action, got %T", result)
	}

	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 See Other, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/protected-page" {
		t.Errorf("Expected redirect to /protected-page, got %q", loc)
	}
}

func TestIssue3095_AccountSwitchingSuccess(t *testing.T) {
	form := url.Values{}
	form.Set("username", "userB")
	form.Set("password", "correcthorse")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	// Pre-populate session as User A (ID = 5)
	session.Values["UID"] = int32(5)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().Add(time.Hour).Unix()

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         20, // User B
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	q.GetLoginRoleForUserFn = func(ctx context.Context, id int32) (int32, error) {
		return 1, nil
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	handler, ok := result.(http.HandlerFunc)
	if !ok {
		t.Fatalf("Expected http.HandlerFunc, got %T", result)
	}
	handler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", rr.Code)
	}

	// Check session is updated to User B
	if uid, ok := session.Values["UID"].(int32); !ok || uid != 20 {
		t.Errorf("Expected session UID to be 20, got %v", session.Values["UID"])
	}
}

func TestIssue3095_AccountSwitchingFailure(t *testing.T) {
	form := url.Values{}
	form.Set("username", "userB")
	form.Set("password", "wrong")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	core.SessionName = "test_session"
	core.Store = sessions.NewCookieStore([]byte("secret"))
	session, _ := core.Store.New(req, core.SessionName)

	// Pre-populate session as User A (ID = 5)
	session.Values["UID"] = int32(5)
	session.Values["LoginTime"] = int64(100)
	session.Values["ExpiryTime"] = int64(200)

	q := testhelpers.NewQuerierStub()
	hash, alg, _ := HashPassword("correcthorse")
	q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
		return &db.SystemGetLoginRow{
			Idusers:         20,
			Passwd:          sql.NullString{String: hash, Valid: true},
			PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
		}, nil
	}
	q.SystemInsertLoginAttemptFn = func(ctx context.Context, params db.SystemInsertLoginAttemptParams) error {
		return nil
	}
	q.GetPasswordResetByUserFn = func(ctx context.Context, arg db.GetPasswordResetByUserParams) (*db.PendingPassword, error) {
		return nil, sql.ErrNoRows
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	rr := httptest.NewRecorder()
	result := loginTask.Action(rr, req)

	// Since login fails, it returns an unexported loginFormHandler. We can just check the string representation or type
	resStr := fmt.Sprintf("%#v", result)
	if !strings.Contains(resStr, "Invalid username or password. You remain logged in as your current account.") {
		t.Errorf("Expected error message indicating A remains logged in, got %v", resStr)
	}

	// Check session A is untouched
	if uid, ok := session.Values["UID"].(int32); !ok || uid != 5 {
		t.Errorf("Expected session UID to remain 5, got %v", session.Values["UID"])
	}
}

func TestIssue3095_LogoutRedirect(t *testing.T) {
	// The logout function is in the user package, so we can't test it directly here without import cycles if we're not careful.
	// We'll test SessionErrorRedirect since it's in core, wait... core is already imported.

	req := httptest.NewRequest("GET", "/some-page", nil)
	rr := httptest.NewRecorder()

	core.RedirectToLogin(rr, req, nil)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 See Other, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Location"), "/login?back=%2Fsome-page") {
		t.Errorf("Expected redirect with encoded back parameter, got %s", rr.Header().Get("Location"))
	}

	if cc := rr.Header().Get("Cache-Control"); cc != "no-cache, no-store, must-revalidate" {
		t.Errorf("Expected Cache-Control header, got %s", cc)
	}
}
T3_EOF

cat << 'T4_EOF' > handlers/auth/loginPage_test.go
package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/sessions"
)

func TestLoginTask_Action(t *testing.T) {
	t.Run("Happy Path - Pending Reset Prompt", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		pwHash, alg, _ := HashPassword("newpw")
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: "oldhash", Valid: true},
				PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetPasswordResetByUserReturns = &db.PendingPassword{
			ID:               2,
			UserID:           1,
			Passwd:           pwHash,
			PasswdAlgorithm:  alg,
			VerificationCode: "code",
			CreatedAt:        time.Now(),
		}

		form := url.Values{"username": {"bob"}, "password": {"newpw"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)

		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		// If SystemInsertLoginAttemptCalls is missing, this will fail compilation.
		// If it is present, it should work.
		if len(q.SystemInsertLoginAttemptCalls) != 0 {
			t.Fatalf("unexpected login attempts recorded: %v", q.SystemInsertLoginAttemptCalls)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "name=\"id\" value=\"2\"") {
			t.Fatalf("missing id field: %q", body)
		}
	})

	t.Run("Happy Path - Signed External Back URL", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 10
		q := testhelpers.NewQuerierStub()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		pwHash, alg, _ := HashPassword("pw")
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: pwHash, Valid: true},
				PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetLoginRoleForUserReturns = 1

		raw := "https://example.org/ok"
		ts := time.Now().Add(time.Hour).Unix()
		sig := SignBackURL("k", raw, ts)
		key := "k"
		form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {raw}, "back_ts": {fmt.Sprint(ts)}, "back_sig": {sig}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		req.Host = "example.com"

		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), q, cfg, common.WithImageSignKey(key), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("status=%d", rr.Code)
		}
		if loc := rr.Header().Get("Location"); loc != raw {
			t.Fatalf("location=%q", loc)
		}
	})

	t.Run("Unhappy Path - No Such User", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetLoginErr = sql.ErrNoRows

		form := url.Values{"username": {"bob"}, "password": {"pw"}}
		req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		ctx := req.Context()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemInsertLoginAttemptCalls) != 1 {
			t.Fatalf("expected login attempt recorded, got %d", len(q.SystemInsertLoginAttemptCalls))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Invalid username or password") {
			t.Fatalf("body=%q", body)
		}
	})

	t.Run("Unhappy Path - Invalid Password", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetPasswordResetByUserErr = sql.ErrNoRows
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: "7c4f29407893c334a6cb7a87bf045c0d", Valid: true},
				PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}

		form := url.Values{"username": {"bob"}, "password": {"wrong"}}
		req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		ctx := req.Context()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemInsertLoginAttemptCalls) != 1 {
			t.Fatalf("expected login attempt recorded, got %d", len(q.SystemInsertLoginAttemptCalls))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Invalid username or password") {
			t.Fatalf("body=%q", body)
		}
	})

	t.Run("Unhappy Path - Invalid Password Preserves Back Data", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.GetPasswordResetByUserErr = sql.ErrNoRows
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: "7c4f29407893c334a6cb7a87bf045c0d", Valid: true},
				PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}

		form := url.Values{
			"username": {"bob"},
			"password": {"wrong"},
			"back":     {"/target"},
			"method":   {http.MethodPost},
		}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		ctx := req.Context()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if len(q.SystemInsertLoginAttemptCalls) != 1 {
			t.Fatalf("expected login attempt recorded, got %d", len(q.SystemInsertLoginAttemptCalls))
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Invalid username or password") {
			t.Fatalf("body=%q", body)
		}
		if !strings.Contains(body, "name=\"back\" value=\"/target\"") {
			t.Fatalf("missing back field: %q", body)
		}
		if !strings.Contains(body, "name=\"method\" value=\"POST\"") {
			t.Fatalf("missing method field: %q", body)
		}
	})

	t.Run("Unhappy Path - External Back URL Ignored", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		pwHash, alg, _ := HashPassword("pw")
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:         1,
				Passwd:          sql.NullString{String: pwHash, Valid: true},
				PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
				Username:        sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetLoginRoleForUserReturns = 1

		form := url.Values{"username": {"bob"}, "password": {"pw"}, "back": {"https://evil.com"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		req.Host = "example.com"

		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Fatalf("status=%d", rr.Code)
		}
		if loc := rr.Header().Get("Location"); loc != "/" {
			t.Fatalf("location=%q", loc)
		}
	})

	t.Run("Unhappy Path - Throttle", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		// Mock SystemCountRecentLoginAttempts to return threshold
		q.SystemCountRecentLoginAttemptsReturns = 5

		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 3
		cfg.LoginAttemptWindow = 15

		form := url.Values{"username": {"bob"}, "password": {"pw"}}
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "1.2.3.4:1111"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)

		cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status=%d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Too many failed attempts") {
			t.Fatalf("body=%q", rr.Body.String())
		}
	})
}

func TestLoginTask_Page(t *testing.T) {
	t.Run("Happy Path - Hidden Fields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login?code=abc&back=%2Ffoo&method=POST&data=x", nil)
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		loginTask.Page(rr, req)

		body := rr.Body.String()
		if !strings.Contains(body, "name=\"code\" value=\"abc\"") {
			t.Fatalf("missing code field: %q", body)
		}
		if !strings.Contains(body, "name=\"back\" value=\"/foo\"") {
			t.Fatalf("missing back field: %q", body)
		}
		if !strings.Contains(body, "name=\"method\" value=\"POST\"") {
			t.Fatalf("missing method field: %q", body)
		}
		if strings.Contains(body, "back_sig") || strings.Contains(body, "back_ts") || strings.Contains(body, "name=\"data\"") {
			t.Fatalf("unexpected signature or data fields: %q", body)
		}
	})

	t.Run("Happy Path - Signed Back URL", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 10
		raw := "https://evil.com/x"
		ts := time.Now().Add(time.Hour).Unix()
		sig := SignBackURL("k", raw, ts)
		req := httptest.NewRequest(http.MethodGet, "/login?back="+url.QueryEscape(raw)+"&back_ts="+fmt.Sprint(ts)+"&back_sig="+sig, nil)
		req.Host = "example.com"
		key := "k"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), cfg, common.WithImageSignKey(key), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		loginTask.Page(rr, req)

		body := rr.Body.String()
		if !strings.Contains(body, "name=\"back\" value=\""+raw+"\"") {
			t.Fatalf("missing back field: %q", body)
		}
	})

	t.Run("Unhappy Path - Invalid Back URL", func(t *testing.T) {
		cfg := config.NewRuntimeConfig()
		cfg.LoginAttemptThreshold = 10
		raw := "https://evil.com/x"
		req := httptest.NewRequest(http.MethodGet, "/login?back="+url.QueryEscape(raw), nil)
		req.Host = "example.com"
		key := "k"
		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"
		session, _ := store.New(req, core.SessionName)
		cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), cfg, common.WithImageSignKey(key), common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		loginTask.Page(rr, req)

		body := rr.Body.String()
		if strings.Contains(body, "name=\"back\"") {
			t.Fatalf("unexpected back field: %q", body)
		}
	})
}

func TestHappyPathLoginFormHandler_ActionTarget(t *testing.T) {
	req := httptest.NewRequest("GET", "/login", nil)
	cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()
	h := loginFormHandler{msg: "foo"}
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "foo") {
		t.Fatalf("missing message")
	}
	if !strings.Contains(body, "action=\"/login\"") {
		t.Fatalf("missing form action")
	}
}

func TestHappyPathSanitizeBackURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig())
	res, _ := cd.SanitizeBackURL(req, "/some/path")
	if res != "/some/path" {
		t.Fatalf("backURL=%s", res)
	}
}

func TestHappyPathSanitizeBackURLSigned(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	key := "test-key"
	cd := common.NewCoreData(req.Context(), testhelpers.NewQuerierStub(), config.NewRuntimeConfig(), common.WithImageSignKey(key))
	raw := "https://evil.com/"
	ts := time.Now().Add(time.Hour).Unix()
	sig := SignBackURL(key, raw, ts)
	req.Form = url.Values{"back_ts": {fmt.Sprint(ts)}, "back_sig": {sig}}
	res, _ := cd.SanitizeBackURL(req, raw)
	if res != raw {
		t.Fatalf("backURL=%s", res)
	}
}

func TestPasskeyLoginUsesUserBoundCeremony(t *testing.T) {
	t.Run("BeginLogin session is validated as a known user", func(t *testing.T) {
		// Mock out DB and user
		q := testhelpers.NewQuerierStub()
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{
				Idusers:  1,
				Username: sql.NullString{String: "bob", Valid: true},
			}, nil
		}
		q.GetPasskeysByUserIDFn = func(ctx context.Context, userID int32) ([]db.Passkey, error) {
			return []db.Passkey{
				{
					CredentialID: []byte("cred-id"),
					PublicKey:    []byte("pub-key"),
				},
			}, nil
		}

		cfg := config.NewRuntimeConfig()
		cfg.SiteDomain = "localhost"

		req := httptest.NewRequest(http.MethodGet, "/login/passkey/begin?username=bob", nil)
		store := sessions.NewCookieStore([]byte("test"))
		session, _ := store.New(req, core.SessionName)

		cd := common.NewCoreData(req.Context(), q, cfg, common.WithSession(session))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		// Execute
		loginPasskeyBegin(rr, req)

		// Assert OK status
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		// Ensure that the session has the WebAuthn user bound (ID 1)
		waUserRaw := session.Values["webauthn_user"]
		if waUserRaw == nil {
			t.Fatalf("expected webauthn_user to be stored in session, but it was nil")
		}

		waUserID, ok := waUserRaw.(int32)
		if !ok || waUserID != 1 {
			t.Fatalf("expected webauthn_user ID to be 1, got %v", waUserRaw)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "publicKey") {
			t.Fatalf("expected publicKey in JSON response, got: %s", body)
		}
	})
}

func TestPasskeyUnavailableResponseDoesNotRevealUserExistence(t *testing.T) {
	t.Run("Missing user and user without passkeys share response", func(t *testing.T) {
		q1 := testhelpers.NewQuerierStub()
		q1.SystemGetLoginErr = sql.ErrNoRows

		cfg := config.NewRuntimeConfig()
		req := httptest.NewRequest(http.MethodGet, "/login/passkey/begin?username=missing", nil)
		store := sessions.NewCookieStore([]byte("test"))
		session, _ := store.New(req, core.SessionName)

		cd1 := common.NewCoreData(req.Context(), q1, cfg, common.WithSession(session))
		ctx1 := context.WithValue(req.Context(), consts.KeyCoreData, cd1)
		req1 := req.WithContext(ctx1)

		rr1 := httptest.NewRecorder()
		loginPasskeyBegin(rr1, req1)

		q2 := testhelpers.NewQuerierStub()
		q2.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			return &db.SystemGetLoginRow{Idusers: 2, Username: sql.NullString{String: "nopasskeys", Valid: true}}, nil
		}
		q2.GetPasskeysByUserIDReturns = nil

		cd2 := common.NewCoreData(req.Context(), q2, cfg, common.WithSession(session))
		ctx2 := context.WithValue(req.Context(), consts.KeyCoreData, cd2)
		req2 := httptest.NewRequest(http.MethodGet, "/login/passkey/begin?username=nopasskeys", nil).WithContext(ctx2)
		rr2 := httptest.NewRecorder()
		loginPasskeyBegin(rr2, req2)

		if rr1.Code != http.StatusNotFound {
			t.Errorf("Missing user: expected 404, got %d", rr1.Code)
		}
		if rr2.Code != http.StatusNotFound {
			t.Errorf("No passkeys user: expected 404, got %d", rr2.Code)
		}
		if rr1.Body.String() != rr2.Body.String() {
			t.Errorf("Responses differed. Missing user got %q, no-passkey user got %q", rr1.Body.String(), rr2.Body.String())
		}
	})
}

func TestLoginTask_Security_UsernameEnumeration(t *testing.T) {
	// Helper to extract error message from response body
	getErrorMsg := func(body string) string {
		if strings.Contains(body, "No such user") {
			return "No such user"
		}
		if strings.Contains(body, "Invalid password") {
			return "Invalid password"
		}
		if strings.Contains(body, "Invalid username or password") {
			return "Invalid username or password"
		}
		return "Unknown error message"
	}

	t.Run("Error messages should be identical", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemGetLoginFn = func(ctx context.Context, username sql.NullString) (*db.SystemGetLoginRow, error) {
			if username.String == "valid_user" {
				return &db.SystemGetLoginRow{
					Idusers:         1,
					Passwd:          sql.NullString{String: "somehash", Valid: true},
					PasswdAlgorithm: sql.NullString{String: "md5", Valid: true},
					Username:        username,
				}, nil
			}
			return nil, sql.ErrNoRows
		}

		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test-session"

		// Test invalid user
		form1 := url.Values{"username": {"invalid_user"}, "password": {"pw"}}
		req1 := httptest.NewRequest("POST", "/login", strings.NewReader(form1.Encode()))
		req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		session1, _ := store.New(req1, core.SessionName)
		cd1 := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithSession(session1))
		ctx1 := context.WithValue(req1.Context(), consts.KeyCoreData, cd1)
		req1 = req1.WithContext(ctx1)
		rr1 := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr1, req1)
		msg1 := getErrorMsg(rr1.Body.String())

		// Test valid user, wrong password
		form2 := url.Values{"username": {"valid_user"}, "password": {"wrong_pw"}}
		req2 := httptest.NewRequest("POST", "/login", strings.NewReader(form2.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		session2, _ := store.New(req2, core.SessionName)
		cd2 := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithSession(session2))
		ctx2 := context.WithValue(req2.Context(), consts.KeyCoreData, cd2)
		req2 = req2.WithContext(ctx2)
		rr2 := httptest.NewRecorder()
		handlers.TaskHandler(loginTask)(rr2, req2)
		msg2 := getErrorMsg(rr2.Body.String())

		if msg1 != msg2 {
			t.Errorf("Error messages differ: invalid_user='%s', valid_user_wrong_pw='%s'", msg1, msg2)
		}
	})
}

func TestLoginTaskTemplatesRequiredExist(t *testing.T) {
	for _, req := range loginTask.RequiredTemplates() {
		if !req.Exists() {
			t.Errorf("Missing template %s", req)
		}
	}
}
T4_EOF

cat << 'R_EOF' > handlers/user/pages_test.go
package user_test

import (
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/user"
)

var allPages = []handlers.Page{
	user.UserAppearancePage,
	user.UserEmailPage,
	user.UserEmailVerifyCodePage,
	user.UserPage,
	user.UserPublicProfilePage,
	user.UserPublicProfileSettingPage,
	user.UserResetPasswordPage,
	user.AdminDashboardPage,
	user.AdminUserApprovePage,
	user.AdminUserEditCommentsPage,
	user.AdminUserEditGrantsPage,
	user.AdminUserEditRolesPage,
	user.AdminUserRenamePage,
	user.AdminRunTaskPage,
	user.AdminUserEditPage,
	user.AdminUserResetPasswordPage,
	user.UserGalleryPage,
	user.UserLangPage,
	user.UserNotificationsPage,
	user.UserNotificationOpenPage,
	user.UserPagingPage,
	user.UserSubscriptionAddPage,
	user.UserSubscriptionsPage,
	user.UserThreadSubscriptionsPage,
	user.UserTimezonePage,
}

func TestPagesExist(t *testing.T) {
	for _, p := range allPages {
		t.Run(string(p), func(t *testing.T) {
			if !p.Exists() {
				t.Fatalf("page template %s missing", string(p))
			}
		})
	}
}
R_EOF

echo "Done writing files"
