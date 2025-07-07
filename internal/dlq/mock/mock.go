package mock

import (
	"context"
	"sync"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Record stores a DLQ message recorded by the Provider.
type Record struct{ Message string }

// Provider records DLQ messages in memory for testing.
type Provider struct {
	mu      sync.Mutex
	Records []Record
}

// Record appends the message to the in-memory slice.
func (p *Provider) Record(_ context.Context, msg string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Records = append(p.Records, Record{Message: msg})
	return nil
}

func providerFromConfig(_ runtimeconfig.RuntimeConfig, _ *dbpkg.Queries) dlq.DLQ {
	return &Provider{}
}

// Register registers the mock provider factory.
func Register() { dlq.RegisterProvider("mock", providerFromConfig) }
