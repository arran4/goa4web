package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	imagesign "github.com/arran4/goa4web/internal/images"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/gorilla/sessions"
)

// handleDie responds with an internal server error.
func handleDie(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
}

// CoreAdderMiddlewareWithDB populates request context with CoreData for
// templates using the supplied database handle. The verbosity controls optional
// logging of database pool statistics.
func CoreAdderMiddlewareWithDB(db *sql.DB, cfg config.RuntimeConfig, verbosity int, emailReg *email.Registry, startTime time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := core.GetSession(r)
			if err != nil {
				core.SessionErrorRedirect(w, r, err)
				return
			}
			var uid int32
			if v, ok := session.Values["UID"].(int32); ok {
				uid = v
			}
			if expi, ok := session.Values["ExpiryTime"]; ok {
				var exp int64
				switch t := expi.(type) {
				case int64:
					exp = t
				case int:
					exp = int64(t)
				case float64:
					exp = int64(t)
				}
				if exp != 0 && time.Now().Unix() > exp {
					delete(session.Values, "UID")
					delete(session.Values, "LoginTime")
					delete(session.Values, "ExpiryTime")
					RedirectToLogin(w, r, session)
					return
				}
			}
			if db == nil {
				ue := common.UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
				log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
				http.Error(w, ue.ErrorMessage, http.StatusInternalServerError)
				return
			}

			queries := dbpkg.New(db)
			sm := queries
			if verbosity > 0 {
				log.Printf("db pool stats: %+v", db.Stats())
			}

			if session.ID != "" {
				if uid != 0 {
					if err := queries.InsertSession(r.Context(), dbpkg.InsertSessionParams{SessionID: session.ID, UsersIdusers: uid}); err != nil {
						log.Printf("insert session: %v", err)
					}
				} else {
					if err := queries.DeleteSessionByID(r.Context(), session.ID); err != nil {
						log.Printf("delete session: %v", err)
					}
				}
			}

			base := "http://" + r.Host
			if cfg.HTTPHostname != "" {
				base = strings.TrimRight(cfg.HTTPHostname, "/")
			}
			provider := emailReg.ProviderFromConfig(cfg)
			cd := common.NewCoreData(r.Context(), queries,
				common.WithImageURLMapper(imagesign.MapURL),
				common.WithSession(session),
				common.WithEmailProvider(provider),
				common.WithAbsoluteURLBase(base),
				common.WithConfig(cfg),
				common.WithStartTime(startTime),
				common.WithSessionManager(sm))
			cd.UserID = uid
			_ = cd.UserRoles()

			idx := nav.IndexItems()
			cd.IndexItems = idx
			cd.Title = "Arran's Site"
			cd.FeedsEnabled = cfg.FeedsEnabled
			cd.AdminMode = r.URL.Query().Get("mode") == "admin"
			if uid != 0 && handlers.NotificationsEnabled() {
				cd.NotificationCount = int32(cd.UnreadNotificationCount())
			}
			ctx := context.WithValue(r.Context(), consts.KeyCoreData, cd)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
		log.Printf("%s %s uid=%d session=%s", r.Method, r.URL.Path, uid, sessID)
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
func RedirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if session != nil {
		if err := session.Save(r, w); err != nil {
			log.Printf("save session: %v", err)
		}
	}
	vals := url.Values{}
	vals.Set("back", r.URL.RequestURI())
	if r.Method != http.MethodGet {
		if err := r.ParseForm(); err == nil {
			vals.Set("method", r.Method)
			vals.Set("data", r.Form.Encode())
		}
	}
	http.Redirect(w, r, "/login?"+vals.Encode(), http.StatusTemporaryRedirect)
}
