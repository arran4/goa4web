package goa4web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/templates"
)

func adminShutdownPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin",
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
