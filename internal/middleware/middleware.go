package middleware

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	nav "github.com/arran4/goa4web/internal/navigation"
	imagesign "github.com/arran4/goa4web/pkg/images"
	"github.com/gorilla/sessions"
)

// handleDie responds with an internal server error.
func handleDie(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
}

// IndexItem exposes the core/common navigation item type.
type IndexItem = common.IndexItem

// CoreAdderMiddleware populates request context with CoreData for templates.
func CoreAdderMiddleware(next http.Handler) http.Handler {
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
				redirectToLogin(w, r, session)
				return
			}
		}

		queries := r.Context().Value(hcommon.KeyQueries).(*dbpkg.Queries)
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

		cd := common.NewCoreData(r.Context(), queries,
			common.WithImageURLMapper(imagesign.MapURL),
			common.WithSession(session))
		cd.UserID = uid
		_ = cd.Roles()

		idx := nav.IndexItems()
		if uid != 0 {
			idx = append(idx, common.IndexItem{Name: "Preferences", Link: "/usr"})
		}
		cd.IndexItems = idx
		cd.Title = "Arran's Site"
		cd.FeedsEnabled = config.AppRuntimeConfig.FeedsEnabled
		cd.AdminMode = r.URL.Query().Get("mode") == "admin"
		if uid != 0 && hcommon.NotificationsEnabled() {
			if c := cd.UnreadNotificationCount(); c > 0 {
				idx = append(idx, common.IndexItem{Name: fmt.Sprintf("Notifications (%d)", c), Link: "/usr/notifications"})
			}
		}
		cd.IndexItems = idx
		ctx := context.WithValue(r.Context(), hcommon.KeyCoreData, cd)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// DBPool should be assigned by the parent package to supply the database.
var DBPool *sql.DB

// DBAdderMiddleware injects the database and queries into the request context.
func DBAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if DBPool == nil {
			ue := common.UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
			log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
			http.Error(w, ue.ErrorMessage, http.StatusInternalServerError)
			return
		}
		if dbLogVerbosity > 0 {
			log.Printf("db pool stats: %+v", DBPool.Stats())
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, hcommon.KeySQLDB, DBPool)
		ctx = context.WithValue(ctx, hcommon.KeyQueries, dbpkg.New(DBPool))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetDBPool configures the database handle and logging verbosity used by
// DBAdderMiddleware.
func SetDBPool(db *sql.DB, verbosity int) {
	DBPool = db
	dbLogVerbosity = verbosity
}

// dbLogVerbosity controls optional logging of database pool stats.
var dbLogVerbosity int

// Configuration stores simple key/value pairs loaded from a file.
type Configuration struct {
	data map[string]string
}

// NewConfiguration creates an empty Configuration.
func NewConfiguration() *Configuration {
	return &Configuration{data: make(map[string]string)}
}

func (c *Configuration) set(key, value string) {
	c.data[key] = value
}

func (c *Configuration) get(key string) string {
	return c.data[key]
}

// readConfiguration populates Configuration from a file on the provided fs.
func (c *Configuration) readConfiguration(fs core.FileSystem, filename string) {
	b, err := fs.ReadFile(filename)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := scanner.Text()
		sep := strings.Index(line, "=")
		if sep >= 0 {
			key := line[:sep]
			value := line[sep+1:]
			c.set(key, value)
		}
	}
}

// X2c converts a two character hex string into a byte.
func X2c(what string) byte {
	digit := func(c byte) byte {
		if c >= 'A' {
			return (c & 0xdf) - 'A' + 10
		}
		return c - '0'
	}

	d1 := digit(what[0])
	d2 := digit(what[1])
	return d1*16 + d2
}

// RequestLoggerMiddleware logs incoming requests along with the user and session IDs.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := int32(0)
		sessID := ""
		if cd, ok := r.Context().Value(hcommon.KeyCoreData).(*common.CoreData); ok && cd != nil {
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
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if session != nil {
		backURL := r.URL.RequestURI()
		session.Values["BackURL"] = backURL
		if r.Method != http.MethodGet {
			if err := r.ParseForm(); err == nil {
				session.Values["BackMethod"] = r.Method
				session.Values["BackData"] = r.Form.Encode()
			}
		} else {
			delete(session.Values, "BackMethod")
			delete(session.Values, "BackData")
		}
		_ = session.Save(r, w)
	}
	http.Redirect(w, r, "/login?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusTemporaryRedirect)
}
