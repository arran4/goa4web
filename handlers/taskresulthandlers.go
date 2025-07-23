package handlers

import "net/http"

type RedirectHandler string

type templateWithDataHandler struct {
	tmpl string
	data any
}

var _ http.Handler = (*templateWithDataHandler)(nil)

func TemplateWithDataHandler(tmpl string, data any) any {
	return &templateWithDataHandler{tmpl: tmpl, data: data}
}

// TemplateHandler renders the template and handles any template error.
// Example usage:
//
// type Data struct{ *CoreData }
// TemplateHandler(w, r, "page.gohtml", Data{cd})
//
// Template helpers are provided via data.CoreData.Funcs(r).
func (th *templateWithDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	TemplateHandler(w, r, th.tmpl, th.data)
}
