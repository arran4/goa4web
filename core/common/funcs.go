package common

import (
	"fmt"
	"github.com/arran4/goa4web"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/csrf"
)

// Funcs returns template helpers configured with cd's ImageURLMapper.
func (cd *CoreData) Funcs(r *http.Request) template.FuncMap {
	// newsCache memoizes LatestNews results for a single template execution.
	var newsCache any
	var LatestWritings any
	mapper := cd.ImageURLMapper
	return map[string]any{
		"cd":        func() *CoreData { return cd },
		"now":       func() time.Time { return time.Now() },
		"csrfField": func() template.HTML { return csrf.TemplateField(r) },
		"csrfToken": func() string { return csrf.Token(r) },
		"version":   func() string { return goa4web.Version },
		"a4code2html": func(s string) template.HTML {
			c := a4code2html.New(mapper)
			c.CodeType = a4code2html.CTHTML
			c.SetInput(s)
			out, err := io.ReadAll(c.Process())
			if err != nil {
				log.Printf("read markup: %v", err)
			}
			if cerr := c.Error(); cerr != nil {
				log.Printf("process markup: %v", cerr)
			}
			return template.HTML(out)
		},
		"a4code2string": func(s string) string {
			c := a4code2html.New(mapper)
			c.CodeType = a4code2html.CTWordsOnly
			c.SetInput(s)
			out, err := io.ReadAll(c.Process())
			if err != nil {
				log.Printf("read markup: %v", err)
			}
			if cerr := c.Error(); cerr != nil {
				log.Printf("process markup: %v", cerr)
			}
			return string(out)
		},
		"firstline": func(s string) string {
			return strings.Split(s, "\n")[0]
		},
		"left": func(i int, s string) string {
			l := len(s)
			if l > i {
				l = i
			}
			return s[:l]
		},
		"int32": func(i any) int32 {
			switch v := i.(type) {
			case int:
				return int32(v)
			case int32:
				return v
			case int64:
				return int32(v)
			case string:
				n, _ := strconv.Atoi(v)
				return int32(n)
			default:
				return 0
			}
		},
		"addmode": func(u string) string {
			cd, _ := r.Context().Value(consts.KeyCoreData).(*CoreData)
			if cd == nil || !cd.AdminMode {
				return u
			}
			if strings.Contains(u, "?") {
				return u + "&mode=admin"
			}
			return u + "?mode=admin"
		},
		"LatestNews": func() (any, error) {
			if newsCache != nil {
				return newsCache, nil
			}
			posts, err := cd.LatestNews(r)
			if err != nil {
				return nil, fmt.Errorf("latestNews: %w", err)
			}
			newsCache = posts
			return posts, nil
		},
		"LatestWritings": func() (any, error) {
			if LatestWritings != nil {
				return LatestWritings, nil
			}
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			wrs, err := cd.LatestWritings(WithWritingsOffset(int32(offset)))
			if err != nil {
				return nil, fmt.Errorf("latestWritings: %w", err)
			}
			LatestWritings = wrs
			return wrs, nil
		},
	}
}
