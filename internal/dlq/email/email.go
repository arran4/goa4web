package email

import (
	"context"
	"fmt"
	"log"
	"net/mail"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
)

// DLQ sends DLQ messages to administrator emails using the configured provider.
type DLQ struct {
	Provider email.Provider
	Queries  *dbpkg.Queries
}

// Record emails the message to the configured recipients.
func (e DLQ) Record(ctx context.Context, message string) error {
	if e.Provider == nil {
		return fmt.Errorf("no email provider")
	}
	fromAddr := email.ParseAddress(config.AppRuntimeConfig.EmailFrom)
	if f, err := mail.ParseAddress(config.AppRuntimeConfig.EmailFrom); err == nil {
		fromAddr = *f
	}
	for _, addrStr := range config.GetAdminEmails(ctx, e.Queries) {
		toAddr := mail.Address{Address: addrStr}
		msg, err := email.BuildMessage(fromAddr, toAddr, "DLQ message", message, "")
		if err != nil {
			log.Printf("build message: %v", err)
			continue
		}
		if err := e.Provider.Send(ctx, toAddr, msg); err != nil {
			log.Printf("dlq email: %v", err)
		}
	}
	return nil
}

// Register registers the email provider.
func Register() {
	dlq.RegisterProvider("email", func(cfg config.RuntimeConfig, q *dbpkg.Queries) dlq.DLQ {
		p := email.ProviderFromConfig(cfg)
		if p == nil {
			return dlq.LogDLQ{}
		}
		return DLQ{Provider: p, Queries: q}
	})
}
