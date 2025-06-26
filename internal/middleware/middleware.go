package middleware

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

// handleDie responds with an internal server error.
func handleDie(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
}

// IndexItem exposes the core/common navigation item type.
type IndexItem = common.IndexItem

// indexItems are always present navigation links.
var indexItems = []common.IndexItem{
	{Name: "News", Link: "/"},
	{Name: "FAQ", Link: "/faq"},
	{Name: "Blogs", Link: "/blogs"},
	{Name: "Forum", Link: "/forum"},
	{Name: "Linker", Link: "/linker"},
	{Name: "Bookmarks", Link: "/bookmarks"},
	{Name: "ImageBBS", Link: "/imagebbs"},
	{Name: "Search", Link: "/search"},
	{Name: "Writings", Link: "/writings"},
	{Name: "Information", Link: "/information"},
}

// CoreAdderMiddleware populates request context with CoreData for templates.
func CoreAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := core.GetSession(r)
		if err != nil {
			core.SessionErrorRedirect(w, r, err)
			return
		}
		var uid int32
		if err == nil {
			uid, _ = session.Values["UID"].(int32)
		}
		queries := r.Context().Value(hcommon.KeyQueries).(*dbpkg.Queries)

		level := "reader"
		if uid != 0 {
			perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(r.Context(), dbpkg.GetPermissionsByUserIdAndSectionAndSectionAllParams{
				UsersIdusers: uid,
				Section:      sql.NullString{String: "all", Valid: true},
			})
			if err == nil && perm.Level.Valid {
				level = perm.Level.String
			}
		}

		idx := make([]common.IndexItem, len(indexItems))
		copy(idx, indexItems)
		if uid != 0 {
			idx = append(idx, common.IndexItem{Name: "Preferences", Link: "/usr"})
		}
		var count int32
		if uid != 0 && hcommon.NotificationsEnabled() {
			c, err := queries.CountUnreadNotifications(r.Context(), uid)
			if err == nil {
				count = c
				idx = append(idx, common.IndexItem{Name: fmt.Sprintf("Notifications (%d)", c), Link: "/usr/notifications"})
			}
		}
		var ann *dbpkg.GetActiveAnnouncementWithNewsRow
		if queries.DB() != nil {
			if a, err := queries.GetActiveAnnouncementWithNews(r.Context()); err == nil {
				ann = a
			}
		}
		cd := &common.CoreData{
			SecurityLevel:     level,
			IndexItems:        idx,
			UserID:            uid,
			Title:             "Arran's Site",
			FeedsEnabled:      runtimeconfig.AppRuntimeConfig.FeedsEnabled,
			NotificationCount: count,
			Announcement:      ann,
		}
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

// routerWrapper wraps a router with additional middleware.
type routerWrapper interface {
	Wrap(http.Handler) http.Handler
}

// routerWrapperFunc allows ordinary functions to satisfy routerWrapper.
type routerWrapperFunc func(http.Handler) http.Handler

func (f routerWrapperFunc) Wrap(h http.Handler) http.Handler { return f(h) }

// newMiddlewareChain returns a routerWrapper that wraps a handler with the provided
// middleware functions in the order supplied.
func newMiddlewareChain(mw ...func(http.Handler) http.Handler) routerWrapper {
	return routerWrapperFunc(func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	})
}

// RequestLoggerMiddleware logs incoming requests and the associated user ID.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid int32
		if u, ok := r.Context().Value(hcommon.KeyUser).(*dbpkg.User); ok && u != nil {
			uid = u.Idusers
		}
		log.Printf("%s %s uid=%d", r.Method, r.URL.Path, uid)
		next.ServeHTTP(w, r)
	})
}
