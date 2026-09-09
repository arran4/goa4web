package handlers

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// TemplateHandler renders the template and handles any template error.
// Example usage:
//
//	type Data struct{ Message string }
//	TemplateHandler(w, r, "page.gohtml", Data{"hello"})
//
// Template helpers are provided via the CoreData in the request context,
// accessible in templates as Funcs["cd"] (*common.CoreData).
func TemplateHandler(w http.ResponseWriter, r *http.Request, tmpl Page, data any) error {
	cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	sessionCookieName := "goa4web_session"
	if cd != nil && cd.Config != nil && cd.Config.SessionName != "" {
		sessionCookieName = cd.Config.SessionName
	}

	_, err := r.Cookie(sessionCookieName)
	hasCookie := err == nil

	if (cd != nil && cd.UserID != 0) || hasCookie {
		DisableCaching(w)
	} else {
		// Explicitly allow caching for purely anonymous, cookie-less requests
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	if err := tmpl.TemplateExecute(w, r, data); err != nil {
		log.Printf("Template Error: %s", err)
		errData := struct {
			Error   string
			BackURL string
		}{
			Error:   err.Error(),
			BackURL: r.Referer(),
		}
		if err2 := TaskErrorAcknowledgementPageTmpl.TemplateExecute(w, r, errData); err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			RenderErrorPage(w, r, common.ErrInternalServerError)
		}
		return err
	}
	return nil
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
