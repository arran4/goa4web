package handlers

import "net/http"

type RedirectHandler string

// TextByteWriter responds with a plain text byte slice.
type TextByteWriter []byte

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
