package dlq

import (
	"context"
)

// MultiDLQ records messages to multiple DLQs.
type MultiDLQ struct{ providers []DLQ }

// NewMulti returns a MultiDLQ using the provided DLQs.
func NewMulti(providers ...DLQ) MultiDLQ {
	return MultiDLQ{providers: providers}
}

// Record writes the message to all providers. It returns the first error encountered.
func (m MultiDLQ) Record(ctx context.Context, msg string) error {
	var first error
	for _, p := range m.providers {
		if err := p.Record(ctx, msg); err != nil && first == nil {
			first = err
		}
	}
	return first
}
