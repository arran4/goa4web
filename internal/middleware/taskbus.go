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
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/eventbus"
)

// TaskEventMiddlewareWithBus records form tasks on the provided event bus after processing.

// statusRecorder captures the response status for later inspection.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
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
	bus       *eventbus.Bus
	queue     *eventQueue
	log       *log.Logger
	dlq       dlq.DLQ
	processor TaskEventProcessor
}

// TaskEventMiddlewareOption customizes TaskEventMiddleware behavior.
type TaskEventMiddlewareOption func(*TaskEventMiddleware)

// TaskEventProcessor handles task events explicitly instead of consuming them via
// an event bus worker.
type TaskEventProcessor interface {
	ProcessEvent(ctx context.Context, evt eventbus.TaskEvent, q dlq.DLQ) error
}

// WithLogger overrides the logger used by TaskEventMiddleware.
func WithLogger(l *log.Logger) TaskEventMiddlewareOption {
	return func(m *TaskEventMiddleware) {
		if l != nil {
			m.log = l
		}
	}
}

// WithDLQ configures optional DLQ reporting for middleware warnings/errors.
func WithDLQ(q dlq.DLQ) TaskEventMiddlewareOption {
	return func(m *TaskEventMiddleware) {
		m.dlq = q
	}
}

// WithTaskEventProcessor configures explicit task event processing.
func WithTaskEventProcessor(p TaskEventProcessor) TaskEventMiddlewareOption {
	return func(m *TaskEventMiddleware) { m.processor = p }
}

// NewTaskEventMiddleware creates a middleware instance using the provided bus.
func NewTaskEventMiddleware(bus *eventbus.Bus, opts ...TaskEventMiddlewareOption) *TaskEventMiddleware {
	if bus == nil {
		panic("TaskEventMiddleware requires a bus")
	}
	m := &TaskEventMiddleware{
		bus:   bus,
		queue: newEventQueue(maxQueuedTaskEvents, bus),
		log:   log.Default(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	return m
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

// Middleware returns a http.Handler middleware that records task events.
func (m *TaskEventMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if sr.status < http.StatusBadRequest && evt.Outcome == "" {
			evt.Outcome = eventbus.TaskOutcomeSuccess
		}
		if sr.status < http.StatusBadRequest {
			if !eventHasTask(evt) {
				if task := strings.TrimSpace(r.PostFormValue("task")); task != "" {
					evt.Task = tasks.TaskString(task)
				}
			}
			if eventHasTask(evt) {
				if m.processor != nil {
					if err := m.processor.ProcessEvent(r.Context(), *evt, m.dlq); err != nil {
						m.reportIssue(r.Context(), "explicit task event processing failed: %v", err)
					}
				}
				if err := m.bus.Publish(*evt); err != nil {
					if err == eventbus.ErrBusClosed {
						m.queue.enqueue(*evt)
					} else {
						m.reportIssue(r.Context(), "publish task event: %v", err)
					}
				}
			} else if isStateChangingMethod(r.Method) {
				m.reportIssue(r.Context(), "TaskEventMiddleware: successful state-changing request without attached task for %s %s", r.Method, r.URL.Path)
			}
		}
		m.queue.flush(r.Context())
	})
}

func (m *TaskEventMiddleware) reportIssue(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	m.log.Print(msg)
	if m.dlq == nil {
		return
	}
	if err := m.dlq.Record(ctx, msg); err != nil {
		m.log.Printf("task middleware dlq record: %v", err)
	}
}

func eventHasTask(evt *eventbus.TaskEvent) bool {
	if evt == nil || evt.Task == nil {
		return false
	}
	named, ok := evt.Task.(tasks.Name)
	if !ok {
		return true
	}
	name := strings.TrimSpace(named.Name())
	return name != "" && name != "MISSING"
}

func isStateChangingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
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

// SetTaskEventProcessor updates the explicit task event processor.
func (m *TaskEventMiddleware) SetTaskEventProcessor(p TaskEventProcessor) {
	m.processor = p
}

// TaskEventMiddlewareWithBus wraps NewTaskEventMiddleware for backward compatibility.
func TaskEventMiddlewareWithBus(bus *eventbus.Bus) func(http.Handler) http.Handler {
	return NewTaskEventMiddleware(bus).Middleware
}
