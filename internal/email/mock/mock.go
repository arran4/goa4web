package mock

import (
	"context"
	"sync"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// SentMail records a delivered email message.
type SentMail struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Provider collects sent messages in memory for testing.
type Provider struct {
	mu       sync.Mutex
	Messages []SentMail
}

// Send appends the message to the Provider's Messages slice.
func (p *Provider) Send(_ context.Context, to, subject, textBody, htmlBody string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Messages = append(p.Messages, SentMail{To: to, Subject: subject, Text: textBody, HTML: htmlBody})
	return nil
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return &Provider{} }

// Register registers the mock provider factory.
func Register() { email.RegisterProvider("mock", providerFromConfig) }
