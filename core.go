package goa4web

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
	"github.com/arran4/goa4web/runtimeconfig"
)

func handleDie(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
}

// IndexItem struct.
type IndexItem = core.IndexItem

// indexItems.
var indexItems = []core.IndexItem{
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

func CoreAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, err := core.GetSession(request)
		if err != nil {
			core.SessionErrorRedirect(writer, request, err)
			return
		}
		var uid int32
		if err == nil {
			uid, _ = session.Values["UID"].(int32)
		}
		queries := request.Context().Value(ContextValues("queries")).(*Queries)

		level := "reader"
		if uid != 0 {
			perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(request.Context(), GetPermissionsByUserIdAndSectionAndSectionAllParams{
				UsersIdusers: uid,
				Section:      sql.NullString{String: "all", Valid: true},
			})
			if err == nil && perm.Level.Valid {
				level = perm.Level.String
			}
		}

		idx := make([]core.IndexItem, len(indexItems))
		copy(idx, indexItems)
		if uid != 0 {
			idx = append(idx, core.IndexItem{Name: "Preferences", Link: "/usr"})
		}
		var count int32
		if uid != 0 && notificationsEnabled() {
			c, err := queries.CountUnreadNotifications(request.Context(), uid)
			if err == nil {
				count = c
				idx = append(idx, core.IndexItem{Name: fmt.Sprintf("Notifications (%d)", c), Link: "/usr/notifications"})
			}
		}
		var ann *GetActiveAnnouncementWithNewsRow
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
			FeedsEnabled:      runtimeconfig.AppRuntimeConfig.FeedsEnabled,
			NotificationCount: count,
			Announcement:      ann,
		})
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

type CoreData = core.CoreData

type Configuration struct {
	data map[string]string
}

func NewConfiguration() *Configuration {
	return &Configuration{
		data: make(map[string]string),
	}
}

func (c *Configuration) set(key, value string) {
	c.data[key] = value
}

func (c *Configuration) get(key string) string {
	return c.data[key]
}

func (c *Configuration) readConfiguration(filename string) {
	b, err := readFile(filename)
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

// ContextValues is an alias to core.ContextValues for backward compatibility.
type ContextValues = core.ContextValues

func DBAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if dbPool == nil {
			ue := UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
			log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
			http.Error(writer, ue.ErrorMessage, http.StatusInternalServerError)
			return
		}
		if dbLogVerbosity > 0 {
			log.Printf("db pool stats: %+v", dbPool.Stats())
		}
		ctx := request.Context()
		ctx = context.WithValue(ctx, ContextValues("sql.DB"), dbPool)
		ctx = context.WithValue(ctx, ContextValues("queries"), New(dbPool))
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// getPageSize returns the preferred page size within configured bounds.
func getPageSize(r *http.Request) int {
	size := runtimeconfig.AppRuntimeConfig.PageSizeDefault
	if pref, _ := r.Context().Value(ContextValues("preference")).(*Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < runtimeconfig.AppRuntimeConfig.PageSizeMin {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMin
	}
	if size > runtimeconfig.AppRuntimeConfig.PageSizeMax {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMax
	}
	return size
}
