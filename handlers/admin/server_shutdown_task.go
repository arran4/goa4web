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
type ServerShutdownTask struct {
	tasks.TaskString
	h *Handlers
}

// NewServerShutdownTask exposes the shutdown task through the task system. This
// is 100% auditable via the audit log.
func (h *Handlers) NewServerShutdownTask() *ServerShutdownTask {
	return &ServerShutdownTask{TaskString: TaskServerShutdown, h: h}
}

var _ tasks.Task = (*ServerShutdownTask)(nil)
var _ tasks.TaskMatcher = (*ServerShutdownTask)(nil)
var _ tasks.TemplatesRequired = (*ServerShutdownTask)(nil)

func (t *ServerShutdownTask) Matcher() mux.MatcherFunc {
	taskM := tasks.HasTask(string(TaskServerShutdown))
	adminM := handlers.RequiredAdminAccess()
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return taskM(r, m) && adminM(r, m)
	}
}

func (t *ServerShutdownTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminAccess() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin",
	}
	path := r.URL.Path
	uid := cd.UserID
	go func() {
		if t.h != nil && t.h.Srv != nil && t.h.Srv.Bus != nil {
			evt := eventbus.TaskEvent{
				Path:    path,
				Task:    TaskServerShutdown,
				UserID:  uid,
				Time:    time.Now(),
				Outcome: eventbus.TaskOutcomeSuccess,
			}
			if err := t.h.Srv.Bus.Publish(evt); err != nil {
				log.Printf("publish shutdown event: %v", err)
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if t.h != nil && t.h.Srv != nil {
			if err := t.h.Srv.Shutdown(ctx); err != nil {
				log.Printf("shutdown error: %v", err)
			}
		}
	}()
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func (t *ServerShutdownTask) TemplatesRequired() []string {
	return []string{handlers.TemplateRunTaskPage}
}
