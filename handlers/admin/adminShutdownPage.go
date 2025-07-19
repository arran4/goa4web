package admin

import (
	"context"
	"log"
	"net/http"
	"time"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
)

func AdminShutdownPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Back:     "/admin",
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := Srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	handlers.TemplateHandler(w, r, "tasks/run_task.gohtml", data)
}
