package emailutil

import (
	"context"
	"fmt"
	"log"
	"time"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// emailQueueWorker periodically sends pending emails using the provided provider.
func EmailQueueWorker(ctx context.Context, q *db.Queries, provider email.Provider, dlqProvider dlq.DLQ, interval time.Duration) {
	if q == nil || provider == nil {
		log.Printf("email queue worker disabled: missing queue or provider")
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ProcessPendingEmail(ctx, q, provider, dlqProvider)
		case <-ctx.Done():
			return
		}
	}
}

// ProcessPendingEmail sends a single queued email if available.
func ProcessPendingEmail(ctx context.Context, q *db.Queries, provider email.Provider, dlqProvider dlq.DLQ) {
	if q == nil || provider == nil {
		return
	}
	if !EmailSendingEnabled() {
		return
	}
	if provider == nil {
		log.Println("No mail provider specified")
		return
	}
	emails, err := q.FetchPendingEmails(ctx, 1)
	if err != nil {
		log.Printf("fetch queue: %v", err)
		return
	}
	if len(emails) == 0 {
		return
	}
	e := emails[0]
	from := email.ParseAddress(runtimeconfig.AppRuntimeConfig.EmailFrom)
	to := email.ParseAddress(e.ToEmail)
	msg, err := email.BuildMessage(from, to, e.Subject, e.Body, "")
	if err != nil {
		log.Printf("build message: %v", err)
		return
	}
	if err := provider.Send(ctx, to.Address, e.Subject, msg); err != nil {
		log.Printf("send queued mail: %v", err)
		count, incErr := q.IncrementEmailError(ctx, e.ID)
		if incErr != nil {
			log.Printf("increment email error: %v", incErr)
			return
		}
		if count > 4 {
			if dlqProvider != nil {
				_ = dlqProvider.Record(ctx, fmt.Sprintf("email %d to %s failed: %v", e.ID, e.ToEmail, err))
			}
			if delErr := q.DeletePendingEmail(ctx, e.ID); delErr != nil {
				log.Printf("delete email: %v", delErr)
			}
		}
		return
	}
	if err := q.MarkEmailSent(ctx, e.ID); err != nil {
		log.Printf("mark sent: %v", err)
	}
}
