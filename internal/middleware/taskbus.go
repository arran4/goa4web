package middleware

import (
	"context"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arran4/goa4web/core/common"
	coreconsts "github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
)

// TaskBus middleware records form tasks on the event bus after processing.

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
	events   []eventbus.TaskEvent
}

func newEventQueue(capacity int) *eventQueue {
	return &eventQueue{capacity: capacity}
}

func (q *eventQueue) enqueue(evt eventbus.TaskEvent) {
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
	events := append([]eventbus.TaskEvent(nil), q.events...)
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

func (q *eventQueue) queued() []eventbus.TaskEvent {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]eventbus.TaskEvent(nil), q.events...)
}

// TaskBus provides the task event middleware and exposes queued events.
type TaskBus struct {
	queue *eventQueue
}

// NewTaskBus returns a TaskBus with an empty queue.
func NewTaskBus() *TaskBus {
	return &TaskBus{queue: newEventQueue(maxQueuedTaskEvents)}
}

// QueuedEvents returns a snapshot of currently buffered events.
func (tb *TaskBus) QueuedEvents() []eventbus.TaskEvent { return tb.queue.queued() }

// Flush publishes all buffered events.
func (tb *TaskBus) Flush(ctx context.Context) { tb.queue.flush(ctx) }

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware records form tasks on the event bus after processing.
func (tb *TaskBus) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := r.PostFormValue("task")
		cd, ok := r.Context().Value(coreconsts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			log.Panicf("TaskBus middleware: missing CoreData for %s", r.URL.Path)
		}
		uid := cd.UserID
		admin := strings.Contains(r.URL.Path, "/admin")
		_ = admin
		evt := &eventbus.TaskEvent{
			Path:   r.URL.Path,
			Task:   tasks.TaskString("MISSING"),
			UserID: uid,
			Time:   time.Now(),
			Data:   map[string]any{},
		}
		cd.SetEvent(evt)
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)
		if task != "" && sr.status < http.StatusBadRequest {
			if err := eventbus.DefaultBus.Publish(*evt); err != nil {
				if err == eventbus.ErrBusClosed {
					tb.queue.enqueue(*evt)
				} else {
					log.Printf("publish task event: %v", err)
				}
			}
		}
		tb.queue.flush(r.Context())
	})
}
