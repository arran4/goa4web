package common

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/templates"
)

// TemplateHandler renders name using the standard template functions.
func TemplateHandler(w http.ResponseWriter, r *http.Request, name string, data any) {
	if err := templates.RenderTemplate(w, name, data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// IndexMiddleware adds custom navigation links before next executes.
func IndexMiddleware(fn func(*CoreData, *http.Request)) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cd, ok := r.Context().Value(KeyCoreData).(*CoreData); ok && fn != nil {
				fn(cd, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}
