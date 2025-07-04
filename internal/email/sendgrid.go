//go:build sendgrid
// +build sendgrid

package email

import (
	"context"
	"fmt"
	"log"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendgridBuilt indicates whether the SendGrid provider is compiled in.
const SendgridBuilt = true

// SendGridProvider sends mail using the SendGrid API.
type SendGridProvider struct{ APIKey string }

func (s SendGridProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	from := mail.NewEmail("", SourceEmail)
	toAddr := mail.NewEmail("", to)
	msg := mail.NewSingleEmail(from, subject, toAddr, textBody, htmlBody)
	client := sendgrid.NewSendClient(s.APIKey)
	resp, err := client.SendWithContext(ctx, msg)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid: status %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

func SendGridProviderFromConfig(key string) Provider {
	if key == "" {
		log.Printf("Email disabled: SENDGRID_KEY not set")
		return nil
	}
	return SendGridProvider{APIKey: key}
}
