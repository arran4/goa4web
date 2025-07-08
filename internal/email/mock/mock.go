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
	Raw     []byte
	Text    string
	HTML    string
}

// Provider collects sent messages in memory for testing.
type Provider struct {
	mu       sync.Mutex
	Messages []SentMail
}

// Send appends the message to the Provider's Messages slice.
func (p *Provider) Send(_ context.Context, to mail.Address, rawEmailMessage []byte) error {
	m, err := mail.ReadMessage(bytes.NewReader(rawEmailMessage))
	var subject, textBody, htmlBody string
	if err != nil {
		// Fallback for plain body with no headers
		textBody = string(rawEmailMessage)
	} else {
		dec := new(mime.WordDecoder)
		subject, _ = dec.DecodeHeader(m.Header.Get("Subject"))
		ct := m.Header.Get("Content-Type")
		mediaType, params, _ := mime.ParseMediaType(ct)
		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(m.Body, params["boundary"])
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}
				b, _ := io.ReadAll(part)
				t, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
				switch t {
				case "text/plain":
					textBody = string(b)
				case "text/html":
					htmlBody = string(b)
				}
			}
		} else if mediaType == "text/plain" {
			b, _ := io.ReadAll(m.Body)
			textBody = string(b)
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Messages = append(p.Messages, SentMail{To: to.Address, Subject: subject, Raw: rawEmailMessage, Text: textBody, HTML: htmlBody})
	return nil
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return &Provider{} }

// Register registers the mock provider factory.
func Register() { email.RegisterProvider("mock", providerFromConfig) }
