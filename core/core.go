package core

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/sessions"
)

// Dependencies that can be set by the main package.
var (
	GetSession           func(*http.Request) (*sessions.Session, error)
	SessionErrorRedirect func(http.ResponseWriter, *http.Request, error)
	NotificationsEnabled func() bool

	FeedsEnabled    bool
	PageSizeDefault int
	PageSizeMin     int
	PageSizeMax     int

	DBPool         *sql.DB
	DBLogVerbosity int
)

// HandleDie writes a 500 response with the provided message.
func HandleDie(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
}

// IndexItem struct.
type IndexItem struct {
	Name string // Name of URL displayed in <a href>
	Link string // URL for link.
}

// indexItems.
var indexItems = []IndexItem{
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

// CoreAdderMiddleware injects core data into the request context.
func CoreAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, err := GetSession(request)
		if err != nil {
			SessionErrorRedirect(writer, request, err)
			return
		}
		var uid int32
		if err == nil {
			uid, _ = session.Values["UID"].(int32)
		}
		queries := request.Context().Value(ContextValues("queries")).(*db.Queries)

		level := "reader"
		if uid != 0 {
			perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(request.Context(), db.GetPermissionsByUserIdAndSectionAndSectionAllParams{
				UsersIdusers: uid,
				Section:      sql.NullString{String: "all", Valid: true},
			})
			if err == nil && perm.Level.Valid {
				level = perm.Level.String
			}
		}

		idx := make([]IndexItem, len(indexItems))
		copy(idx, indexItems)
		if uid != 0 {
			idx = append(idx, IndexItem{Name: "Preferences", Link: "/usr"})
		}
		var count int32
		if uid != 0 && NotificationsEnabled != nil && NotificationsEnabled() {
			c, err := queries.CountUnreadNotifications(request.Context(), uid)
			if err == nil {
				count = c
				idx = append(idx, IndexItem{Name: fmt.Sprintf("Notifications (%d)", c), Link: "/usr/notifications"})
			}
		}
		var ann *db.GetActiveAnnouncementWithNewsRow
		if queries.DB() != nil {
			if a, err := queries.GetActiveAnnouncementWithNews(request.Context()); err == nil {
				ann = a
			}
		}
		ctx := context.WithValue(request.Context(), ContextValues("coreData"), &CoreData{
			SecurityLevel:     level,
			IndexItems:        idx,
			UserID:            uid,
			Title:             "Arran4's Website",
			FeedsEnabled:      FeedsEnabled,
			NotificationCount: count,
			Announcement:      ann,
		})
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// CoreData holds common values available to templates.
type CoreData struct {
	IndexItems        []IndexItem
	CustomIndexItems  []IndexItem
	UserID            int32
	SecurityLevel     string
	Title             string
	AutoRefresh       bool
	FeedsEnabled      bool
	RSSFeedUrl        string
	AtomFeedUrl       string
	NotificationCount int32
	Announcement      *db.GetActiveAnnouncementWithNewsRow
}

func (cd *CoreData) GetPermissionsByUserIdAndSectionAndSectionAll() string {
	return cd.SecurityLevel
}

var rolePriority = map[string]int{
	"reader":        1,
	"writer":        2,
	"moderator":     3,
	"administrator": 4,
}

// HasRole reports if the current user has at least the specified role.
func (cd *CoreData) HasRole(role string) bool {
	return rolePriority[cd.SecurityLevel] >= rolePriority[role]
}

type Configuration struct {
	data map[string]string
}

// NewConfiguration creates an empty Configuration.
func NewConfiguration() *Configuration {
	return &Configuration{
		data: make(map[string]string),
	}
}

func (c *Configuration) Set(key, value string) {
	c.data[key] = value
}

func (c *Configuration) Get(key string) string {
	return c.data[key]
}

// ReadFile is a helper that can be replaced by tests.
var ReadFile = os.ReadFile

func (c *Configuration) ReadConfiguration(filename string) {
	b, err := ReadFile(filename)
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
			c.Set(key, value)
		}
	}
}

// X2c converts two hex digits to a byte.
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

type ContextValues string

// DBAdderMiddleware injects database handles into the request context.
func DBAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if DBPool == nil {
			ue := UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
			log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
			http.Error(writer, ue.ErrorMessage, http.StatusInternalServerError)
			return
		}
		if DBLogVerbosity > 0 {
			log.Printf("db pool stats: %+v", DBPool.Stats())
		}
		ctx := request.Context()
		ctx = context.WithValue(ctx, ContextValues("sql.DB"), DBPool)
		ctx = context.WithValue(ctx, ContextValues("queries"), db.New(DBPool))
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// getPageSize returns the preferred page size within configured bounds.
func GetPageSize(r *http.Request) int {
	size := PageSizeDefault
	if pref, _ := r.Context().Value(ContextValues("preference")).(*Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < PageSizeMin {
		size = PageSizeMin
	}
	if size > PageSizeMax {
		size = PageSizeMax
	}
	return size
}

// Preference mirrors a user preference with only the fields we need here.
type Preference struct {
	PageSize int32
}

// UserError mirrors goa4web.UserError for internal use.
type UserError struct {
	Err          error
	ErrorMessage string
}
