//go:build sendgrid
// +build sendgrid

package sendgrid

import (
	"context"
	"fmt"
	"log"

	sg "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Built indicates whether the SendGrid provider is compiled in.
const Built = true

// Provider sends mail using the SendGrid API.
type Provider struct {
	APIKey string
	From   string
}

func (s Provider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	from := mail.NewEmail("", s.From)
	toAddr := mail.NewEmail("", to)
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

func providerFromConfig(key string, from string) email.Provider {
	if key == "" {
		log.Printf("Email disabled: SENDGRID_KEY not set")
		return nil
	}
	return Provider{APIKey: key, From: from}
}

// Register registers the SendGrid provider factory.
func Register() {
	email.RegisterProvider("sendgrid", func(cfg runtimeconfig.RuntimeConfig) email.Provider {
		return providerFromConfig(cfg.EmailSendGridKey, cfg.EmailFrom)
	})
}
