package handlers

import "net/http"

type RedirectHandler string

// templateWithDataHandler is a small wrapper that renders tmpl with the
// provided data when ServeHTTP is called.
type templateWithDataHandler struct {
	tmpl string
	data any
}

var _ http.Handler = (*templateWithDataHandler)(nil)

// TemplateWithDataHandler returns an http.Handler that renders tmpl with data
// using TemplateHandler. It is useful for returning templates from tasks.
func TemplateWithDataHandler(tmpl string, data any) any {
	return &templateWithDataHandler{tmpl: tmpl, data: data}
}

func (th *templateWithDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	TemplateHandler(w, r, th.tmpl, th.data)
}
