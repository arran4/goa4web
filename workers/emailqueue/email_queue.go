package emailqueue

import (
	"context"
	"fmt"
	"github.com/arran4/goa4web/config"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
)

// EmailQueueWorker sends pending emails ensuring a minimum delay between sends.
// When a bus is provided the worker wakes up immediately when a new message is
// queued by listening for EmailQueueEvent messages.
func EmailQueueWorker(ctx context.Context, q db.Querier, provider email.Provider, dlqProvider dlq.DLQ, bus *eventbus.Bus, delay time.Duration, cfg *config.RuntimeConfig) {
	if q == nil {
		log.Printf("email queue worker disabled: queue configured=%v", q != nil)
		return
	}
	var ch <-chan eventbus.Message
	if bus != nil {
		ch = bus.Subscribe(eventbus.EmailQueueMessageType)
	}
	for {
		if ProcessPendingEmail(ctx, q, provider, dlqProvider, cfg) {
			log.Printf("email queue worker: waiting %s", delay)
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

// adminBypassAddr extracts the recipient address from the email body and
// returns it when it matches one of the configured administrator emails.
func adminBypassAddr(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig, body string) (mail.Address, bool) {
	m, err := mail.ReadMessage(strings.NewReader(body))
	if err != nil {
		return mail.Address{}, false
	}
	addr, err := mail.ParseAddress(m.Header.Get("To"))
	if err != nil {
		return mail.Address{}, false
	}
	for _, a := range config.GetAdminEmails(ctx, q, cfg) {
		if strings.EqualFold(a, addr.Address) {
			return *addr, true
		}
	}
	return mail.Address{}, false
}

func isAdminEmail(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig, addr string) bool {
	for _, a := range config.GetAdminEmails(ctx, q, cfg) {
		if strings.EqualFold(a, addr) {
			return true
		}
	}
	return false
}

func hasVerificationRecord(ctx context.Context, q db.Querier, addr string) bool {
	if q == nil {
		return false
	}
	ue, err := q.GetUserEmailByEmail(ctx, addr)
	if err != nil {
		return false
	}
	if ue.VerifiedAt.Valid {
		return false
	}
	if ue.VerificationExpiresAt.Valid && ue.VerificationExpiresAt.Time.Before(time.Now()) {
		return false
	}
	return ue.LastVerificationCode.Valid
}

// ResolveQueuedEmailAddress resolves the recipient for a queued email.
// When the user record is missing or lacks a valid address the admin or direct
// email logic is applied.
func ResolveQueuedEmailAddress(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig, e *db.SystemListPendingEmailsRow) (mail.Address, error) {
	if e.ToUserID.Valid && e.ToUserID.Int32 != 0 {
		user, err := q.SystemGetUserByID(ctx, e.ToUserID.Int32)
		if err == nil && user.Email.Valid && user.Email.String != "" {
			return mail.Address{Name: user.Username.String, Address: user.Email.String}, nil
		}
		if err != nil {
			return mail.Address{}, fmt.Errorf("get user: %v", err)
		}
	}

	m, err := mail.ReadMessage(strings.NewReader(e.Body))
	if err != nil {
		return mail.Address{}, fmt.Errorf("parse message: %v", err)
	}
	addr, err := mail.ParseAddress(m.Header.Get("To"))
	if err != nil {
		return mail.Address{}, fmt.Errorf("parse address: %v", err)
	}

	if isAdminEmail(ctx, q, cfg, addr.Address) {
		addr.Name = "Admin"
		return *addr, nil
	}

	if e.DirectEmail {
		if hasVerificationRecord(ctx, q, addr.Address) {
			return *addr, nil
		}
		return mail.Address{}, fmt.Errorf("no verification record for %s", addr.Address)
	}

	if e.ToUserID.Valid {
		return mail.Address{}, fmt.Errorf("invalid email for user %d", e.ToUserID.Int32)
	}
	return mail.Address{}, fmt.Errorf("unknown recipient")
}

// ProcessPendingEmail sends a single queued email if available.
func ProcessPendingEmail(ctx context.Context, q db.Querier, provider email.Provider, dlqProvider dlq.DLQ, cfg *config.RuntimeConfig) bool {
	if q == nil {
		return false
	}
	if !cfg.EmailEnabled {
		return false
	}
	emails, err := q.SystemListPendingEmails(ctx, db.SystemListPendingEmailsParams{Limit: 1, Offset: 0})
	if err != nil {
		log.Printf("fetch queue: %v", err)
		return false
	}
	if len(emails) == 0 {
		return false
	}
	e := emails[0]
	addr, err := ResolveQueuedEmailAddress(ctx, q, cfg, e)
	if err != nil {
		log.Printf("ResolveQueuedEmailAddress: %v", err)
		if err := q.SystemIncrementPendingEmailError(ctx, e.ID); err != nil {
			log.Printf("increment email error: %v", err)
		}
		return false
	}
	if provider == nil {
		log.Printf("email provider not configured: cannot send email %d to %s", e.ID, addr.Address)
		if err := q.SystemIncrementPendingEmailError(ctx, e.ID); err != nil {
			log.Printf("increment email error: %v", err)
		}
		count, _ := q.GetPendingEmailErrorCount(ctx, e.ID)
		if count > 4 {
			if dlqProvider != nil {
				msg := fmt.Sprintf("email %d to %s failed: no provider configured\n%s", e.ID, addr.Address, e.Body)
				_ = dlqProvider.Record(ctx, msg)
			}
			if delErr := q.AdminDeletePendingEmail(ctx, e.ID); delErr != nil {
				log.Printf("delete email: %v", delErr)
			}
		}
		return true
	}
	if err := provider.Send(ctx, addr, []byte(e.Body)); err != nil {
		log.Printf("send queued mail: %v", err)
		if err := q.SystemIncrementPendingEmailError(ctx, e.ID); err != nil {
			log.Printf("increment email error: %v", err)
			return true
		}
		count, _ := q.GetPendingEmailErrorCount(ctx, e.ID)
		if count > 4 {
			if dlqProvider != nil {
				msg := fmt.Sprintf("email %d to %s failed: %v\n%s", e.ID, addr.Address, err, e.Body)
				_ = dlqProvider.Record(ctx, msg)
			}
			if delErr := q.AdminDeletePendingEmail(ctx, e.ID); delErr != nil {
				log.Printf("delete email: %v", delErr)
			}
		}
		return true
	}
	if err := q.SystemMarkPendingEmailSent(ctx, e.ID); err != nil {
		log.Printf("mark sent: %v", err)
	}
	return true
}
