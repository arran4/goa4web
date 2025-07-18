package eventbus

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/arran4/goa4web/internal/tasks"
)

// Event represents a task or notification that occurred in the application.
type Event struct {
	Path   string         // Path or URI describing the event source
	Task   tasks.Task     // Task that triggered the event
	UserID int32          // ID of the user performing the action
	Time   time.Time      // Time the event occurred
	Data   map[string]any // Optional template data associated with the event
	Admin  bool           // True when the action occurred within /admin
}

// Bus provides a simple publish/subscribe mechanism for events.
type Bus struct {
	mu          sync.RWMutex
	subscribers []chan Event
	closed      bool
}

// ErrBusClosed is returned when publishing to a bus after Shutdown.
var ErrBusClosed = errors.New("event bus closed")

// NewBus creates an empty bus instance.
func NewBus() *Bus {
	return &Bus{}
}

// Subscribe registers a new subscriber and returns a channel for events.
func (b *Bus) Subscribe() <-chan Event {
	ch := make(chan Event, 1)
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()
	return ch
}

// Publish dispatches an event to all current subscribers.
// It returns ErrBusClosed when publishing after Shutdown.
func (b *Bus) Publish(evt Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}
	subs := append([]chan Event(nil), b.subscribers...)
	b.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- evt:
		default:
		}
	}
	return nil
}

const drainInterval = 10 * time.Millisecond // wait time between draining checks

// Shutdown waits for all queued events to be processed and
// prevents any new events from being published.
func (b *Bus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	b.closed = true
	subs := append([]chan Event(nil), b.subscribers...)
	b.mu.Unlock()
	for _, ch := range subs {
		for {
			if len(ch) == 0 {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(drainInterval)
			}
		}
	}
	return nil
}

var (
	// DefaultBus is the global event bus used across the application.
	DefaultBus = NewBus()
)

// ReopenDefaultBus creates a new DefaultBus instance. Callers should publish
// any queued events to the returned bus once subscribers are registered.
func ReopenDefaultBus() {
	DefaultBus = NewBus()
}
