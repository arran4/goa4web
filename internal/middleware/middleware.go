package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web"
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
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
			uid = cd.UserID
			if s := cd.Session(); s != nil {
				sessID = s.ID
			}
		}
		if !(r.URL.Path == "/ws/notifications" && uid == 0) {
			log.Printf("%s %s uid=%d session=%s", r.Method, r.URL.Path, uid, sessID)
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
	if session != nil {
		if err := session.Save(r, w); err != nil {
			log.Printf("save session: %v", err)
		}
	}
	vals := url.Values{}
	vals.Set("back", r.URL.RequestURI())
	if r.Method != http.MethodGet {
		vals.Set("method", r.Method)
		if err := r.ParseForm(); err == nil {
			vals.Set("data", r.PostForm.Encode())
		} else {
			log.Printf("parse form: %v", err)
		}
	}
	http.Redirect(w, r, "/login?"+vals.Encode(), http.StatusSeeOther)
	return http.StatusSeeOther
}
