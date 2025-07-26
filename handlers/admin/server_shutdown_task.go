package admin

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
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
var _ tasks.TaskMatcher = (*ServerShutdownTask)(nil)

func (ServerShutdownTask) Matcher() mux.MatcherFunc {
	taskM := tasks.HasTask(string(TaskServerShutdown))
	adminM := handlers.RequiredAccess("administrator")
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return taskM(r, m) && adminM(r, m)
	}
}

func (ServerShutdownTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasRole("administrator") {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin",
	}
	path := r.URL.Path
	uid := cd.UserID
	go func() {
		if Srv != nil && Srv.Bus != nil {
			evt := eventbus.TaskEvent{
				Path:   path,
				Task:   TaskServerShutdown,
				UserID: uid,
				Time:   time.Now(),
			}
			if err := Srv.Bus.Publish(evt); err != nil {
				log.Printf("publish shutdown event: %v", err)
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := Srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
}
