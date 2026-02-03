package notifications

import (
	"context"

	"github.com/arran4/goa4web/internal/eventbus"
)

// DigestConsumer consumes digest run events from the bus.
type DigestConsumer struct {
	notifier *Notifier
}

// NewDigestConsumer creates a new digest consumer.
func NewDigestConsumer(n *Notifier) *DigestConsumer {
	return &DigestConsumer{notifier: n}
}

// Run starts the consumer loop.
func (c *DigestConsumer) Run(ctx context.Context) {
	if c.notifier == nil || c.notifier.Bus == nil {
		return
	}
	ch := c.notifier.Bus.Subscribe(eventbus.DigestRunMessageType)
	for {
		select {
		case msg := <-ch:
			if evt, ok := msg.(eventbus.DigestRunEvent); ok {
				c.notifier.ProcessDigestForTime(ctx, evt.Time)
			}
		case <-ctx.Done():
			return
		}
	}
}
