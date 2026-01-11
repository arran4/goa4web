package eventbus

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/arran4/goa4web/internal/tasks"
)

// MessageType identifies the kind of event carried by a Message.
// Values are enumerated using iota.
type MessageType int

const (
	// TaskMessageType identifies task events on the bus.
	TaskMessageType MessageType = iota
	// EmailQueueMessageType identifies email queue events on the bus.
	EmailQueueMessageType
)

// Message represents an item sent over the event bus.
// Different message types are defined below.
type Message interface {
	Type() MessageType
}

// TaskEvent represents a task or notification that occurred in the application.
type TaskEvent struct {
	Path    string         // Path or URI describing the event source
	Task    tasks.Task     // Name of the action/task performed
	UserID  int32          // ID of the user performing the action
	Time    time.Time      // Time the event occurred
	Data    map[string]any // Optional template data associated with the event
	Outcome string         // Outcome describes the result of the task run
}

const (
	// TaskOutcomeSuccess indicates the task completed without error.
	TaskOutcomeSuccess = "success"
)

// Type implements the Message interface.
func (TaskEvent) Type() MessageType { return TaskMessageType }

// EmailQueueEvent notifies the email worker that new mail is queued.
type EmailQueueEvent struct {
	Time time.Time // Time the event was published
}

// Type implements the Message interface.
func (EmailQueueEvent) Type() MessageType { return EmailQueueMessageType }

// Bus provides a simple publish/subscribe mechanism for events.
type subscriber struct {
	ch    chan Message
	types map[MessageType]struct{}
}

// Bus provides a simple publish/subscribe mechanism for events.
type Bus struct {
	mu          sync.RWMutex
	subscribers []subscriber
	closed      bool
	SyncPublish func(Message) // Optional hook for synchronous delivery (mostly for tests)
}

// ErrBusClosed is returned when publishing to a bus after Shutdown.
var ErrBusClosed = errors.New("event bus closed")

// NewBus creates an empty bus instance.
func NewBus() *Bus {
	return &Bus{}
}

// Subscribe registers a new subscriber for the provided message types.
// If no types are supplied the subscriber receives all messages.
func (b *Bus) Subscribe(types ...MessageType) <-chan Message {
	ch := make(chan Message, 1)
	set := make(map[MessageType]struct{}, len(types))
	for _, t := range types {
		set[t] = struct{}{}
	}
	b.mu.Lock()
	b.subscribers = append(b.subscribers, subscriber{ch: ch, types: set})
	b.mu.Unlock()
	return ch
}

// Publish dispatches an event to all current subscribers.
// It returns ErrBusClosed when publishing after Shutdown.
func (b *Bus) Publish(msg Message) error {
	b.mu.RLock()
	syncPub := b.SyncPublish
	b.mu.RUnlock()
	if syncPub != nil {
		syncPub(msg)
	}

	if evt, ok := msg.(TaskEvent); ok {
		if n, ok := evt.Task.(tasks.Name); ok && n.Name() == "MISSING" {
			log.Printf("event bus received MISSING task for path %s", evt.Path)
		}
	}
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}
	subs := append([]subscriber(nil), b.subscribers...)
	b.mu.RUnlock()
	for _, s := range subs {
		if len(s.types) > 0 {
			if _, ok := s.types[msg.Type()]; !ok {
				continue
			}
		}
		select {
		case s.ch <- msg:
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
	subs := append([]subscriber(nil), b.subscribers...)
	b.mu.Unlock()
	for _, s := range subs {
		ch := s.ch
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
