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
	From     mail.Address
}

// Record emails the message to the configured recipients.
func (e DLQ) Record(ctx context.Context, message string) error {
	if e.Provider == nil {
		return fmt.Errorf("no email provider")
	}
	fromAddr := e.From
	if fromAddr.Address == "" {
		fromAddr = email.ParseAddress("")
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
func Register(r *dlq.Registry) {
	r.RegisterProvider("email", func(cfg config.RuntimeConfig, q *dbpkg.Queries) dlq.DLQ {
		p := email.ProviderFromConfig(cfg)
		if p == nil {
			return dlq.LogDLQ{}
		}
		from := email.ParseAddress(cfg.EmailFrom)
		if f, err := mail.ParseAddress(cfg.EmailFrom); err == nil {
			from = *f
		}
		return DLQ{Provider: p, Queries: q, From: from}
	})
}
