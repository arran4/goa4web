package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/eventbus"
)

// TaskEventMiddleware records form tasks on the event bus after processing.

// statusRecorder captures the response status for later inspection.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func TaskEventMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := r.PostFormValue("task")
		uid := int32(0)
		cd, _ := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData)
		if cd != nil {
			uid = cd.UserID
		}
		admin := strings.Contains(r.URL.Path, "/admin")
		evt := &eventbus.Event{
			Path:   r.URL.Path,
			Task:   task,
			UserID: uid,
			Time:   time.Now(),
			Admin:  admin,
		}
		if cd != nil {
			cd.SetEvent(evt)
		}
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)
		if task != "" && sr.status < http.StatusBadRequest {
			if err := eventbus.DefaultBus.Publish(*evt); err != nil {
				log.Printf("publish task event: %v", err)
				// TODO: queue task events for resumption when the bus is closed
			}
		}
	})
}
