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

// Subscription represents an active subscription to the event bus.
type Subscription struct {
	bus       *Bus
	ch        chan Envelope
	types     map[MessageType]struct{}
	filter    func(Message) bool
	reliable  bool
	mu        sync.Mutex
	closed    bool
	done      chan struct{}
	deliverWg sync.WaitGroup
	closeOnce sync.Once
	drainOnce sync.Once
}

// Channel returns the read-only channel of Envelopes for this subscription.
func (s *Subscription) Channel() <-chan Envelope {
	return s.ch
}

// closeDelivery stops any new deliveries, unblocks pending deliver calls,
// and safely closes s.ch so range consumers can finish.
func (s *Subscription) closeDelivery() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		close(s.done)
		s.mu.Unlock()

		if s.bus != nil {
			s.bus.removeSubscriber(s)
		}

		// Wait for all in-flight deliver calls to finish
		s.deliverWg.Wait()

		// Safely close the channel now that no further deliveries can occur
		close(s.ch)
	})
}

// Close unregisters the subscription, closes delivery, and drains/acknowledges
// any remaining unconsumed envelopes so Bus accounting completes cleanly.
func (s *Subscription) Close() {
	s.closeDelivery()

	// Drain any remaining buffered envelopes and acknowledge them
	s.drainOnce.Do(func() {
		for env := range s.ch {
			env.Ack()
		}
	})
}

func (s *Subscription) matches(msg Message) bool {
	if len(s.types) > 0 {
		if _, ok := s.types[msg.Type()]; !ok {
			return false
		}
	}
	if s.filter != nil {
		return s.filter(msg)
	}
	return true
}

func (s *Subscription) deliver(env Envelope) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSubscriptionClosed
	}
	s.deliverWg.Add(1)
	done := s.done
	s.mu.Unlock()

	defer s.deliverWg.Done()

	if s.reliable {
		select {
		case s.ch <- env:
			return nil
		case <-done:
			return ErrSubscriptionClosed
		}
	}

	select {
	case s.ch <- env:
		return nil
	default:
		return ErrBufferFull
	}
}

// Bus provides a publish/subscribe mechanism for events.
type Bus struct {
	mu          sync.RWMutex
	subscribers []*Subscription
	closed      bool
	wg          sync.WaitGroup
	SyncPublish func(Message) // Optional hook for synchronous delivery (mostly for tests)
}

// ErrBusClosed is returned when publishing to a bus after Shutdown.
var ErrBusClosed = errors.New("event bus closed")

// ErrSubscriptionClosed is returned when attempting to deliver to a closed subscription.
var ErrSubscriptionClosed = errors.New("subscription closed")

// ErrBufferFull is returned when a non-reliable subscriber channel is full.
var ErrBufferFull = errors.New("subscriber buffer full")

// NewBus creates an empty bus instance.
func NewBus() *Bus {
	return &Bus{}
}

// SubscribeConfig holds subscription configuration options.
type SubscribeConfig struct {
	Types      []MessageType
	BufferSize int
	Reliable   bool
	Filter     func(Message) bool
	Context    context.Context
}

// SubscribeOption configures a SubscribeConfig.
type SubscribeOption func(*SubscribeConfig)

// WithTypes sets the message types the subscription will receive.
func WithTypes(types ...MessageType) SubscribeOption {
	return func(c *SubscribeConfig) {
		c.Types = append(c.Types, types...)
	}
}

// WithBufferSize sets the channel buffer capacity for the subscription.
func WithBufferSize(size int) SubscribeOption {
	return func(c *SubscribeConfig) {
		if size > 0 {
			c.BufferSize = size
		}
	}
}

// WithReliableDelivery enables reliable delivery with backpressure on full buffer.
func WithReliableDelivery() SubscribeOption {
	return func(c *SubscribeConfig) {
		c.Reliable = true
	}
}

// WithFilter sets a message filter predicate for the subscription.
func WithFilter(fn func(Message) bool) SubscribeOption {
	return func(c *SubscribeConfig) {
		c.Filter = fn
	}
}

// WithContext associates a context with the subscription; when the context is cancelled,
// the subscription is automatically closed.
func WithContext(ctx context.Context) SubscribeOption {
	return func(c *SubscribeConfig) {
		c.Context = ctx
	}
}

// Subscribe registers a new subscriber for the provided message types with default lossy semantics.
// It returns a read-only channel of Envelopes. Consumers must call Ack() on each envelope.
func (b *Bus) Subscribe(types ...MessageType) <-chan Envelope {
	sub := b.SubscribeWithOptions(WithTypes(types...))
	return sub.Channel()
}

// SubscribeWithOptions registers a subscription with fine-grained configuration.
func (b *Bus) SubscribeWithOptions(opts ...SubscribeOption) *Subscription {
	cfg := &SubscribeConfig{
		BufferSize: 100,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	set := make(map[MessageType]struct{}, len(cfg.Types))
	for _, t := range cfg.Types {
		set[t] = struct{}{}
	}

	sub := &Subscription{
		bus:      b,
		ch:       make(chan Envelope, cfg.BufferSize),
		types:    set,
		filter:   cfg.Filter,
		reliable: cfg.Reliable,
		done:     make(chan struct{}),
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		sub.Close()
		return sub
	}
	b.subscribers = append(b.subscribers, sub)
	b.mu.Unlock()

	if cfg.Context != nil {
		go func() {
			select {
			case <-cfg.Context.Done():
				sub.Close()
			case <-sub.done:
			}
		}()
	}

	return sub
}

func (b *Bus) removeSubscriber(s *Subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, sub := range b.subscribers {
		if sub == s {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			break
		}
	}
}

// Publish dispatches an event to all current subscribers.
// It returns ErrBusClosed when publishing after Shutdown.
// If a reliable subscriber fails to accept the event (e.g. cancelled/closed during backpressure),
// Publish returns a non-nil error.
func (b *Bus) Publish(msg Message) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}
	syncPub := b.SyncPublish

	var matching []*Subscription
	for _, s := range b.subscribers {
		if s.matches(msg) {
			matching = append(matching, s)
		}
	}

	if len(matching) == 0 {
		b.mu.RUnlock()
		if syncPub != nil {
			syncPub(msg)
		}
		return nil
	}

	// Register all accepted work under lock before Shutdown can observe closed state
	b.wg.Add(len(matching))
	b.mu.RUnlock()

	if syncPub != nil {
		syncPub(msg)
	}

	if evt, ok := msg.(TaskEvent); ok {
		if n, ok := evt.Task.(tasks.Name); ok && n.Name() == "MISSING" {
			log.Printf("event bus received MISSING task for path %s", evt.Path)
		}
	}

	var firstErr error
	for _, s := range matching {
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

		if err := s.deliver(env); err != nil {
			ack()
			if s.reliable && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Shutdown closes all subscriptions, waits for all queued/in-flight events to be acknowledged,
// and prevents new events from being published.
func (b *Bus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	b.closed = true
	subs := make([]*Subscription, len(b.subscribers))
	copy(subs, b.subscribers)
	b.mu.Unlock()

	for _, s := range subs {
		s.closeDelivery()
	}

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
