package eventbus

import (
	"sync"
	"time"
)

// Event represents a task or notification that occurred in the application.
type Event struct {
	Path   string    // Path or URI describing the event source
	Task   string    // Name of the action/task performed
	UserID int32     // ID of the user performing the action
	Time   time.Time // Time the event occurred
	Item   any       // Item impacted by the action
	Admin  bool      // True when the event came from an admin action
}

// Bus provides a simple publish/subscribe mechanism for events.
type Bus struct {
	mu          sync.RWMutex
	subscribers []chan Event
}

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
func (b *Bus) Publish(evt Event) {
	b.mu.RLock()
	subs := append([]chan Event(nil), b.subscribers...)
	b.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- evt:
		default:
		}
	}
}

var (
	// DefaultBus is the global event bus used across the application.
	DefaultBus = NewBus()
)
