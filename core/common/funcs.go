package common

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/a4code2html"
	"github.com/gorilla/csrf"
)

var Version = "dev"

// NewFuncs returns template helpers for the current request context.
// Deprecated: prefer (*CoreData).Funcs.
func NewFuncs(r *http.Request) template.FuncMap {
	cd, _ := r.Context().Value(ContextValues("coreData")).(*CoreData)
	if cd == nil {
		cd = &CoreData{}
	}
	return cd.Funcs(r)
}

// Funcs returns template helpers configured with cd's ImageURLMapper.
func (cd *CoreData) Funcs(r *http.Request) template.FuncMap {
	// newsCache memoizes LatestNews results for a single template execution.
	var newsCache any
	var LatestWritings any
	mapper := cd.ImageURLMapper
	return map[string]any{
		"now":       func() time.Time { return time.Now() },
		"csrfField": func() template.HTML { return csrf.TemplateField(r) },
		"version":   func() string { return Version },
		"a4code2html": func(s string) template.HTML {
			c := a4code2html.New(mapper)
			c.CodeType = a4code2html.CTHTML
			c.SetInput(s)
			out, _ := io.ReadAll(c.Process())
			return template.HTML(out)
		},
		"a4code2string": func(s string) string {
			c := a4code2html.New(mapper)
			c.CodeType = a4code2html.CTWordsOnly
			c.SetInput(s)
			out, _ := io.ReadAll(c.Process())
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
		"addmode": func(u string) string {
			cd, _ := r.Context().Value(ContextValues("coreData")).(*CoreData)
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
			wrs, err := cd.LatestWritings(r)
			if err != nil {
				return nil, fmt.Errorf("latestWritings: %w", err)
			}
			LatestWritings = wrs
			return wrs, nil
		},
	}
}
