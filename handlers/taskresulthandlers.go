package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RedirectHandler Preserves HTTP method probably wrong for the majority of uses.
type RedirectHandler string

// RefreshDirectHandler should be used over RedirectHandler for most of goa4web's tasks as the router is METHOD aware.
type RefreshDirectHandler struct {
	// TargetURL where to refresh direct to
	TargetURL string
	// Duration time to wait (default 1 sec)
	Duration time.Duration
}

func (rdh RefreshDirectHandler) Content() string {
	data := []string{}
	if rdh.Duration == 0 {
		data = append(data, "1")
	} else {
		data = append(data, strconv.Itoa(int(rdh.Duration/time.Second)))
	}
	if rdh.TargetURL != "" {
		data = append(data, "url="+rdh.TargetURL)
	}
	return strings.Join(data, "; ")
}

// TextByteWriter responds with a plain text byte slice.
type TextByteWriter []byte

// templateWithDataHandler is a small wrapper that renders tmpl with the
// provided data when ServeHTTP is called.
type templateWithDataHandler struct {
	tmpl Page
	data any
}

var _ http.Handler = (*templateWithDataHandler)(nil)

// TemplateWithDataHandler returns an http.Handler that renders tmpl with data
// using TemplateHandler. It is useful for returning templates from tasks.
func TemplateWithDataHandler(tmpl Page, data any) any {
	return &templateWithDataHandler{tmpl: tmpl, data: data}
}

// TemplateHandler renders the template and handles any template error.
// Example usage:
//
//	type Data struct{ *CoreData }
//	TemplateHandler(w, r, "page.gohtml", Data{cd})
//
// Template helpers are provided via data.CoreData.Funcs(r).
func (th *templateWithDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	TemplateHandler(w, r, th.tmpl, th.data)
}
