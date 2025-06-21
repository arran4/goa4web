package main

import (
	"context"
	"log"
	"time"
)

// emailQueueWorker periodically sends pending emails using the provided provider.
func emailQueueWorker(ctx context.Context, q *Queries, provider MailProvider, interval time.Duration) {
	if q == nil || provider == nil {
		log.Printf("email queue worker disabled: missing queue or provider")
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			processPendingEmail(ctx, q, provider)
		case <-ctx.Done():
			return
		}
	}
}

// processPendingEmail sends a single queued email if available.
func processPendingEmail(ctx context.Context, q *Queries, provider MailProvider) {
	if q == nil || provider == nil {
		return
	}
	if !emailSendingEnabled() {
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
	if err := provider.Send(ctx, e.ToEmail, e.Subject, e.Body); err != nil {
		log.Printf("send queued mail: %v", err)
		return
	}
	if err := q.MarkEmailSent(ctx, e.ID); err != nil {
		log.Printf("mark sent: %v", err)
	}
}
