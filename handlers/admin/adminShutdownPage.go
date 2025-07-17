package admin

import (
	"context"
	"log"
	"net/http"
	"time"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
)

func AdminShutdownPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
		Back:     "/admin",
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := Srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	common.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}
