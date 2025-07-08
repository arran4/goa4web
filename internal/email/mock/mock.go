package mock

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
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

func parseRawEmail(raw []byte) (string, string) {
	m, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return string(raw), ""
	}
	ctype := m.Header.Get("Content-Type")
	med, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		b, _ := io.ReadAll(m.Body)
		return string(b), ""
	}
	if strings.HasPrefix(med, "multipart/") {
		mr := multipart.NewReader(m.Body, params["boundary"])
		var textBody, htmlBody string
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			b, _ := io.ReadAll(p)
			ct := p.Header.Get("Content-Type")
			if strings.HasPrefix(ct, "text/plain") {
				textBody = string(b)
			} else if strings.HasPrefix(ct, "text/html") {
				htmlBody = string(b)
			}
		}
		return textBody, htmlBody
	}
	b, _ := io.ReadAll(m.Body)
	return string(b), ""
}

// Send appends the message to the Provider's Messages slice.
func (p *Provider) Send(_ context.Context, to, subject string, rawEmailMessage []byte) error {
	textBody, htmlBody := parseRawEmail(rawEmailMessage)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Messages = append(p.Messages, SentMail{To: to, Subject: subject, Text: textBody, HTML: htmlBody})
	return nil
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return &Provider{} }

// Register registers the mock provider factory.
func Register() { email.RegisterProvider("mock", providerFromConfig) }
