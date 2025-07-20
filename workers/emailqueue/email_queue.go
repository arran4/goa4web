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
	"github.com/arran4/goa4web/internal/eventbus"
)

// EmailQueueWorker sends pending emails ensuring a minimum delay between sends.
// When a bus is provided the worker wakes up immediately when a new message is
// queued by listening for EmailQueueEvent messages.
func EmailQueueWorker(ctx context.Context, q *db.Queries, provider email.Provider, dlqProvider dlq.DLQ, bus *eventbus.Bus, delay time.Duration) {
	if q == nil || provider == nil {
		log.Printf("email queue worker disabled: missing queue or provider")
		return
	}
	var ch <-chan eventbus.Message
	if bus != nil {
		ch = bus.Subscribe(eventbus.EmailQueueMessageType)
	}
	for {
		sent := ProcessPendingEmail(ctx, q, provider, dlqProvider)
		if sent {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
			continue
		}

		if ch == nil {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		} else {
			select {
			case <-ch:
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		}
	}
}

// ProcessPendingEmail sends a single queued email if available.
func ProcessPendingEmail(ctx context.Context, q *db.Queries, provider email.Provider, dlqProvider dlq.DLQ) bool {
	if q == nil || provider == nil {
		return false
	}
	if !config.EmailSendingEnabled() {
		return false
	}
	if provider == nil {
		log.Println("No mail provider specified")
		return false
	}
	emails, err := q.FetchPendingEmails(ctx, 1)
	if err != nil {
		log.Printf("fetch queue: %v", err)
		return false
	}
	if len(emails) == 0 {
		return false
	}
	e := emails[0]
	user, err := q.GetUserById(ctx, e.ToUserID)
	if err != nil {
		log.Printf("get user: %v", err)
		return false
	}
	if !user.Email.Valid || user.Email.String == "" {
		log.Printf("invalid email for user %d", e.ToUserID)
		return false
	}
	addr := mail.Address{Name: user.Username.String, Address: user.Email.String}
	if err := provider.Send(ctx, addr, []byte(e.Body)); err != nil {
		log.Printf("send queued mail: %v", err)
		if err := q.IncrementEmailError(ctx, e.ID); err != nil {
			log.Printf("increment email error: %v", err)
			return true
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
		return true
	}
	if err := q.MarkEmailSent(ctx, e.ID); err != nil {
		log.Printf("mark sent: %v", err)
	}
	return true
}
