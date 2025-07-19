package handlers

import (
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

// TemplateHandler renders the template and handles any template error.
// Example usage:
//
//	type Data struct{ *CoreData }
//	TemplateHandler(w, r, "page.gohtml", Data{cd})
//
// Template helpers are provided via data.CoreData.Funcs(r).
func TemplateHandler(w http.ResponseWriter, r *http.Request, tmpl string, data any) {
	cd, _ := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	if cd == nil {
		cd = &common.CoreData{}
	}
	if err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, tmpl, data); err != nil {
		log.Printf("Template Error: %s", err)
		errData := struct {
			*common.CoreData
			Error   string
			BackURL string
		}{
			CoreData: cd,
			Error:    err.Error(),
			BackURL:  r.Referer(),
		}
		if err2 := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "tasks/error_ack.gohtml", errData); err2 != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// IndexMiddleware injects custom index items via fn before executing the next handler.
func IndexMiddleware(fn func(*common.CoreData, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok && cd != nil {
				fn(cd, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}
