package middleware

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core/common"
	coreconsts "github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/eventbus"
)

// TaskEventMiddlewareWithBus records form tasks on the provided event bus after processing.

// statusRecorder captures the response status for later inspection.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// Hijack delegates to the underlying ResponseWriter when available.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

// maxQueuedTaskEvents limits the number of task events stored while the event
// bus is closed.
const maxQueuedTaskEvents = 100

// eventQueue stores events in memory until they can be published.
type eventQueue struct {
	mu       sync.Mutex
	capacity int
	events   []eventbus.TaskEvent
	bus      *eventbus.Bus
}

func newEventQueue(capacity int, bus *eventbus.Bus) *eventQueue {
	return &eventQueue{capacity: capacity, bus: bus}
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
		bus := q.bus
		if bus == nil {
			q.mu.Lock()
			q.events = append(events[i:], q.events...)
			q.mu.Unlock()
			return
		}
		if err := bus.Publish(e); err != nil {
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

// TaskEventMiddleware provides middleware for publishing task events.
type TaskEventMiddleware struct {
	bus   *eventbus.Bus
	queue *eventQueue
}

// NewTaskEventMiddleware creates a middleware instance using the provided bus.
func NewTaskEventMiddleware(bus *eventbus.Bus) *TaskEventMiddleware {
	if bus == nil {
		panic("TaskEventMiddleware requires a bus")
	}
	return &TaskEventMiddleware{
		bus:   bus,
		queue: newEventQueue(maxQueuedTaskEvents, bus),
	}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware returns a http.Handler middleware that records task events.
func (m *TaskEventMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := r.PostFormValue("task")
		cd, ok := r.Context().Value(coreconsts.KeyCoreData).(*common.CoreData)
		if !ok || cd == nil {
			log.Panicf("TaskEventMiddleware: missing CoreData for %s", r.URL.Path)
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
			if err := m.bus.Publish(*evt); err != nil {
				if err == eventbus.ErrBusClosed {
					m.queue.enqueue(*evt)
				} else {
					log.Printf("publish task event: %v", err)
				}
			}
		}
		m.queue.flush(r.Context())
	})
}

// Events returns a copy of the currently queued events.
func (m *TaskEventMiddleware) Events() []eventbus.TaskEvent {
	m.queue.mu.Lock()
	defer m.queue.mu.Unlock()
	return append([]eventbus.TaskEvent(nil), m.queue.events...)
}

// Flush publishes any queued events to the underlying bus.
func (m *TaskEventMiddleware) Flush(ctx context.Context) {
	m.queue.flush(ctx)
}

// SetBus updates the bus used for publishing and flushing events.
func (m *TaskEventMiddleware) SetBus(bus *eventbus.Bus) {
	m.bus = bus
	m.queue.bus = bus
}

// TaskEventMiddlewareWithBus wraps NewTaskEventMiddleware for backward compatibility.
func TaskEventMiddlewareWithBus(bus *eventbus.Bus) func(http.Handler) http.Handler {
	return NewTaskEventMiddleware(bus).Middleware
}
