package admin

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// TaskServerShutdown stops the HTTP server.
const TaskServerShutdown tasks.TaskString = "Server shutdown"

// ServerShutdownTask gracefully shuts down the running server.
type ServerShutdownTask struct{ tasks.TaskString }

// serverShutdownTask exposes the shutdown task through the task system.
// This is 100% auditable via the audit log.
var serverShutdownTask = &ServerShutdownTask{TaskString: TaskServerShutdown}

var _ tasks.Task = (*ServerShutdownTask)(nil)

func (ServerShutdownTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin",
	}
	go func() {
		// TODO add to bus
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := Srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
