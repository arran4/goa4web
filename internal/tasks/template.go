package tasks

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

// Template represents a template file path.
type Template string

func (t Template) String() string {
	return string(t)
}

// Global hooks to keep tasks package independent of core/common logic while still providing methods on Page.
var (
	Handle          func(w http.ResponseWriter, r *http.Request, t Template, data any) error
	TemplateExecute func(w http.ResponseWriter, r *http.Request, t Template, data any) error
	Handler         func(t Template, data any) http.Handler
)

func (t Template) Handle(w http.ResponseWriter, r *http.Request, data any) error {
	if Handle == nil {
		return fmt.Errorf("template handler not initialized")
	}
	return Handle(w, r, t, data)
}

func (t Template) TemplateExecute(w http.ResponseWriter, r *http.Request, data any) error {
	if TemplateExecute == nil {
		return fmt.Errorf("template executor not initialized")
	}
	return TemplateExecute(w, r, t, data)
}

func (t Template) Exists(opts ...templates.Option) bool {
	return templates.AnyTemplateExists(string(t), opts...)
}

func (t Template) Handler(data any) http.Handler {
	if Handler == nil {
		panic("template handler factory not initialized")
	}
	return Handler(t, data)
}
