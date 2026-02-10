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
	// DigestRunMessageType identifies a scheduled digest run.
	DigestRunMessageType
)

// Message represents an item sent over the event bus.
// Different message types are defined below.
type Message interface {
	Type() MessageType
}

// Envelope wraps a Message with an acknowledgement function.
type Envelope struct {
	Msg Message
	ack func()
}

// Ack signals that the message has been processed.
// It is safe to call multiple times; subsequent calls are no-ops.
func (e *Envelope) Ack() {
	if e.ack != nil {
		e.ack()
	}
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

// DigestRunEvent notifies that a digest run is scheduled for a specific time.
type DigestRunEvent struct {
	Time time.Time
}

// Type implements the Message interface.
func (DigestRunEvent) Type() MessageType { return DigestRunMessageType }

// Bus provides a simple publish/subscribe mechanism for events.
type subscriber struct {
	ch    chan Envelope
	types map[MessageType]struct{}
}

// Bus provides a simple publish/subscribe mechanism for events.
type Bus struct {
	mu          sync.RWMutex
	subscribers []subscriber
	closed      bool
	wg          sync.WaitGroup
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
// It returns a read-only channel of Envelopes. Consumers must call Ack() on each envelope.
func (b *Bus) Subscribe(types ...MessageType) <-chan Envelope {
	ch := make(chan Envelope, 1)
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
	defer b.mu.RUnlock()
	if b.closed {
		return ErrBusClosed
	}
	for _, s := range b.subscribers {
		if len(s.types) > 0 {
			if _, ok := s.types[msg.Type()]; !ok {
				continue
			}
		}

		b.wg.Add(1)

		// Create a separate once per subscriber/message to properly handle drop/send
		var once sync.Once
		ack := func() {
			once.Do(func() {
				b.wg.Done()
			})
		}

		env := Envelope{
			Msg: msg,
			ack: ack,
		}

		select {
		case s.ch <- env:
		default:
			// If channel is full, we drop but must decrease WG immediately
			// effectively auto-acking the dropped message.
			ack()
		}
	}
	return nil
}

// Shutdown waits for all queued events to be processed and
// prevents any new events from being published.
func (b *Bus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	b.closed = true
	// Close all subscriber channels to signal consumers to stop.
	for _, s := range b.subscribers {
		close(s.ch)
	}
	b.mu.Unlock()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
