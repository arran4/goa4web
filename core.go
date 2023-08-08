package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func handleDie(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<b><font color=red>You encountered an error: "+message+"....</font></b>")
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

		ctx := context.WithValue(request.Context(), ContextValues("coreData"), &CoreData{
			SecurityLevel: "administrator",
			IndexItems:    indexItems,
			UserID:        1,
		})
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

type CoreData struct {
	IndexItems    []IndexItem
	UserID        int
	SecurityLevel string
	Title         string
}

func (cd *CoreData) GetSecurityLevel() string {
	return cd.SecurityLevel
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
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
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
		db, err := sql.Open("mysql", "a4web:a4web@tcp(localhost:3306)/a4web")
		if err != nil {
			log.Printf("error sql init: %w", err)
			http.Error(writer, "ERROR", 500)
			return
		}
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				log.Printf("Error closing db: %s", err)
			}
		}(db)
		ctx := request.Context()
		ctx = context.WithValue(ctx, ContextValues("sql.DB"), db)
		ctx = context.WithValue(ctx, ContextValues("queries"), New(db))
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
