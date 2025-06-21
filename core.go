package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handleDie(w http.ResponseWriter, message string) {
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
	{Name: "Preferences", Link: "/user"},
}

func CoreAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, err := GetSession(request)
		if err != nil {
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		uid, _ := session.Values["UID"].(int32)
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

		ctx := context.WithValue(request.Context(), ContextValues("coreData"), &CoreData{
			SecurityLevel: level,
			IndexItems:    indexItems,
			UserID:        uid,
			Title:         "Arran4's Website",
			FeedsEnabled:  FeedsEnabled,
		})
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

type CoreData struct {
	IndexItems       []IndexItem
	CustomIndexItems []IndexItem
	UserID           int32
	SecurityLevel    string
	Title            string
	AutoRefresh      bool
	FeedsEnabled     bool
	RSSFeedUrl       string
	AtomFeedUrl      string
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

func (cd *CoreData) HasRole(role string) bool {
	return rolePriority[cd.SecurityLevel] >= rolePriority[role]
}

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

type ContextValues string

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
