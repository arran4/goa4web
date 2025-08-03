package handlers

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// TemplateHandler renders the template and handles any template error.
// Example usage:
//
//	type Data struct{}
//	TemplateHandler(w, r, "page.gohtml", Data{})
//
// CoreData helpers are available through the "cd" template function.
// Pass a dedicated data struct if the template needs additional fields.
func TemplateHandler(w http.ResponseWriter, r *http.Request, tmpl string, data any) {
	cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.ExecuteSiteTemplate(w, r, tmpl, data); err != nil {
		log.Printf("Template Error: %s", err)
		errData := struct {
			Error   string
			BackURL string
		}{
			Error:   err.Error(),
			BackURL: r.Referer(),
		}
		if err2 := cd.ExecuteSiteTemplate(w, r, "taskErrorAcknowledgementPage.gohtml", errData); err2 != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// IndexMiddleware injects custom index items via fn before executing the next handler.
func IndexMiddleware(fn func(*common.CoreData, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok && cd != nil {
				fn(cd, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}
