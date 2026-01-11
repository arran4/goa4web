package tasks

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

// Page represents a template file path.
type Page string

func (p Page) String() string {
	return string(p)
}

// Global hooks to keep tasks package independent of core/common logic while still providing methods on Page.
var (
	Handle          func(w http.ResponseWriter, r *http.Request, p Page, data any) error
	TemplateExecute func(w http.ResponseWriter, r *http.Request, p Page, data any) error
	Handler         func(p Page, data any) http.Handler
)

func (p Page) Handle(w http.ResponseWriter, r *http.Request, data any) error {
	if Handle == nil {
		return fmt.Errorf("page handler not initialized")
	}
	return Handle(w, r, p, data)
}

func (p Page) TemplateExecute(w http.ResponseWriter, r *http.Request, data any) error {
	if TemplateExecute == nil {
		return fmt.Errorf("template executor not initialized")
	}
	return TemplateExecute(w, r, p, data)
}

func (p Page) Exists(opts ...templates.Option) bool {
	return templates.TemplateExists(string(p), opts...)
}

func (p Page) Handler(data any) http.Handler {
	if Handler == nil {
		panic("page handler factory not initialized")
	}
	return Handler(p, data)
}
