package common

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/csrf"
)

// Funcs returns template helpers configured with cd's ImageURLMapper.
func (cd *CoreData) Funcs(r *http.Request) template.FuncMap {
	mapper := cd.ImageURLMapper
	return map[string]any{
		"cd":        func() *CoreData { return cd },
		"now":       func() time.Time { return time.Now().In(cd.Location()) },
		"csrfField": func() template.HTML { return csrf.TemplateField(r) },
		"csrfToken": func() string { return csrf.Token(r) },
		"version":   func() string { return goa4web.Version },
		"dict": func(values ...any) map[string]any {
			m := make(map[string]any)
			for i := 0; i+1 < len(values); i += 2 {
				k, _ := values[i].(string)
				m[k] = values[i+1]
			}
			return m
		},
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
		"trim":      strings.TrimSpace,
		"localTime": func(t time.Time) time.Time { return t.In(cd.Location()) },
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
		"add": func(a, b int) int { return a + b },
		"since": func(prev, curr time.Time) string {
			if prev.IsZero() {
				return ""
			}
			diff := curr.Sub(prev)
			if diff < 0 {
				diff = -diff
			}
			switch {
			case diff < time.Minute:
				return fmt.Sprintf("%d seconds after last comment", int(diff.Seconds()))
			case diff < time.Hour:
				return fmt.Sprintf("%d minutes after last comment", int(diff.Minutes()))
			case diff < 24*time.Hour:
				return fmt.Sprintf("%d hours after last comment", int(diff.Hours()))
			default:
				return fmt.Sprintf("%d days after last comment", int(diff.Hours()/24))
			}
		},
		"addmode": func(u string) string {
			cd, _ := r.Context().Value(consts.KeyCoreData).(*CoreData)
			if cd == nil || !cd.AdminMode {
				return u
			}
			if parsed, err := url.Parse(u); err == nil {
				if parsed.Path == "/admin" || strings.HasPrefix(parsed.Path, "/admin/") {
					return u
				}
			}
			if strings.Contains(u, "?") {
				return u + "&mode=admin"
			}
			return u + "?mode=admin"
		},
		"LatestNews": func() (any, error) {
			posts, err := cd.LatestNews(r)
			if err != nil {
				return nil, fmt.Errorf("latestNews: %w", err)
			}
			return posts, nil
		},
	}
}
