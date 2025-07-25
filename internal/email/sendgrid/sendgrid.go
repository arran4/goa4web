//go:build sendgrid
// +build sendgrid

package sendgrid

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	sg "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Built indicates whether the SendGrid provider is compiled in.
const Built = true

// Provider sends mail using the SendGrid API.
type Provider struct {
	APIKey string
	From   string
}

func (s Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	subject, textBody, htmlBody := parseRawEmail(rawEmailMessage)
	from := mail.NewEmail("", s.From)
	toAddr := mail.NewEmail(to.Name, to.Address)
	msg := mail.NewSingleEmail(from, subject, toAddr, textBody, htmlBody)
	client := sg.NewSendClient(s.APIKey)
	resp, err := client.SendWithContext(ctx, msg)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid: status %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

func parseRawEmail(raw []byte) (string, string, string) {
	m, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return "", string(raw), ""
	}
	subject := m.Header.Get("Subject")
	ctype := m.Header.Get("Content-Type")
	med, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		b, _ := io.ReadAll(m.Body)
		return subject, string(b), ""
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
		return subject, textBody, htmlBody
	}
	b, _ := io.ReadAll(m.Body)
	return subject, string(b), ""
}

func providerFromConfig(key string, from string) email.Provider {
	if key == "" {
		log.Printf("Email disabled: SENDGRID_KEY not set")
		return nil
	}
	return Provider{APIKey: key, From: from}
}

// Register registers the SendGrid provider factory.
func Register(r *email.Registry) {
	r.RegisterProvider("sendgrid", func(cfg config.RuntimeConfig) email.Provider {
		return providerFromConfig(cfg.EmailSendGridKey, cfg.EmailFrom)
	})
}
