package common

import (
	"bytes"
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
	"github.com/arran4/goa4web/core/templates"
	csrfmiddleware "github.com/arran4/goa4web/internal/middleware/csrf"
)

const (
	// defaultTopicTitle is used when a topic is missing a title.
	defaultTopicTitle = "🧵 Untitled topic"
	// defaultTopicDescription is used when a topic is missing a description.
	defaultTopicDescription = "ℹ️ No description provided"
)

// Funcs returns template helpers configured with cd's ImageURLMapper.
func (cd *CoreData) Funcs(r *http.Request) template.FuncMap {
	mapper := cd.ImageURLMapper

	// Color assignment state for quotes
	assignedColors := make(map[string]int)
	counts := make([]int, 6)

	getColor := func(name string) string {
		if idx, ok := assignedColors[name]; ok {
			return fmt.Sprintf("quote-color-%d", idx)
		}

		// Calculate hash
		h := 0
		for _, c := range name {
			h += int(c)
		}
		pref := h % 6

		best := pref
		// Check for collision with preference
		// If preferred color is already used by someone else (count > 0), try to find an unused color or less used one
		// Only check counts, as "used for an existing name" implies we track usage.
		if counts[pref] > 0 {
			// Try to find an unused color
			foundUnused := false
			for i := 0; i < 6; i++ {
				idx := (pref + i) % 6
				if counts[idx] == 0 {
					best = idx
					foundUnused = true
					break
				}
			}

			// If all used, find the one with minimum usage to ensure even spread
			if !foundUnused {
				minC := counts[pref]
				for i := 0; i < 6; i++ {
					if counts[i] < minC {
						minC = counts[i]
						best = i
					}
				}
			}
		}

		assignedColors[name] = best
		counts[best]++
		return fmt.Sprintf("quote-color-%d", best)
	}

	return map[string]any{
		"cd":        func() *CoreData { return cd },
		"now":       func() time.Time { return time.Now().In(cd.Location()) },
		"localTime": cd.LocalTime,
		"formatLocalTime": func(t time.Time) string {
			return cd.FormatLocalTime(t)
		},
		"csrfField": func() template.HTML { return csrfmiddleware.TemplateField(r) },
		"csrfToken": func() string { return csrfmiddleware.Token(r) },
		"version":   func() string { return goa4web.Version },
		"dict": func(values ...any) map[string]any {
			m := make(map[string]any)
			for i := 0; i+1 < len(values); i += 2 {
				k, _ := values[i].(string)
				m[k] = values[i+1]
			}
			return m
		},
		"highlightSearch": func(s string) template.HTML {
			return HighlightSearchTerms(s, cd.SearchWords())
		},
		"a4code2html": func(s string) template.HTML {
			c := a4code2html.New(mapper, getColor)
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
		"topicTitleOrDefault": func(title string) string {
			if trimmed := strings.TrimSpace(title); trimmed != "" {
				return trimmed
			}
			return defaultTopicTitle
		},
		"topicDescriptionOrDefault": func(description string) string {
			if trimmed := strings.TrimSpace(description); trimmed != "" {
				return trimmed
			}
			return defaultTopicDescription
		},
		"trim": strings.TrimSpace,
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
		"include": func(name string, data any) (template.HTML, error) {
			var buf bytes.Buffer
			t := templates.GetCompiledSiteTemplates(cd.Funcs(r))
			err := t.ExecuteTemplate(&buf, name, data)
			return template.HTML(buf.String()), err
		},
	}
}
