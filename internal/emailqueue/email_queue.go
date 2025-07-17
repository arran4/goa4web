package emailqueue

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/config"
	"log"
	"net/mail"
	"time"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
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
	if !config.EmailSendingEnabled() {
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
	user, err := q.GetUserById(ctx, e.ToUserID)
	if err != nil {
		log.Printf("get user: %v", err)
		return
	}
	if !user.Email.Valid || user.Email.String == "" {
		log.Printf("invalid email for user %d", e.ToUserID)
		return
	}
	addr := mail.Address{Name: user.Username.String, Address: user.Email.String}
	if err := provider.Send(ctx, addr, []byte(e.Body)); err != nil {
		log.Printf("send queued mail: %v", err)
		if err := q.IncrementEmailError(ctx, e.ID); err != nil {
			log.Printf("increment email error: %v", err)
			return
		}
		count, _ := q.GetPendingEmailErrorCount(ctx, e.ID)
		if count > 4 {
			if dlqProvider != nil {
				msg := fmt.Sprintf("email %d to %s failed: %v\n%s", e.ID, user.Email.String, err, e.Body)
				_ = dlqProvider.Record(ctx, msg)
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
