//go:build sendgrid
// +build sendgrid

package goa4web

import (
	"context"
	"fmt"
	"log"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// sendgridBuilt indicates whether the SendGrid provider is compiled in.
const sendgridBuilt = true

// sendGridProvider sends mail using the SendGrid API.
type sendGridProvider struct{ apiKey string }

func (s sendGridProvider) Send(ctx context.Context, to, subject, body string) error {
	from := mail.NewEmail("", SourceEmail)
	toAddr := mail.NewEmail("", to)
	msg := mail.NewSingleEmail(from, subject, toAddr, body, body)
	client := sendgrid.NewSendClient(s.apiKey)
	resp, err := client.SendWithContext(ctx, msg)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid: status %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

func sendGridProviderFromConfig(cfg RuntimeConfig) MailProvider {
	key := cfg.EmailSendGridKey
	if key == "" {
		log.Printf("Email disabled: SENDGRID_KEY not set")
		return nil
	}
	return sendGridProvider{apiKey: key}
}
