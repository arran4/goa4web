package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
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

// maxQueuedTaskEvents limits the number of task events stored while the event
// bus is closed.
const maxQueuedTaskEvents = 100

// eventQueue stores events in memory until they can be published.
type eventQueue struct {
	mu       sync.Mutex
	capacity int
	events   []eventbus.Event
}

func newEventQueue(capacity int) *eventQueue {
	return &eventQueue{capacity: capacity}
}

func (q *eventQueue) enqueue(evt eventbus.Event) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.events) >= q.capacity {
		q.events = append(q.events[1:], evt)
	} else {
		q.events = append(q.events, evt)
	}
}

func (q *eventQueue) flush(ctx context.Context) {
	q.mu.Lock()
	if len(q.events) == 0 {
		q.mu.Unlock()
		return
	}
	events := append([]eventbus.Event(nil), q.events...)
	q.events = nil
	q.mu.Unlock()
	for i, e := range events {
		if ctx.Err() != nil {
			q.mu.Lock()
			q.events = append(events[i:], q.events...)
			q.mu.Unlock()
			return
		}
		if err := eventbus.DefaultBus.Publish(e); err != nil {
			if err == eventbus.ErrBusClosed {
				q.mu.Lock()
				q.events = append(events[i:], q.events...)
				q.mu.Unlock()
				return
			}
			log.Printf("flush queued events: %v", err)
		}
	}
}

var taskQueue = newEventQueue(maxQueuedTaskEvents)

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
				if err == eventbus.ErrBusClosed {
					taskQueue.enqueue(*evt)
				} else {
					log.Printf("publish task event: %v", err)
				}
			}
		}
		taskQueue.flush(r.Context())
	})
}
