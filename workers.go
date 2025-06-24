package goa4web

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// startWorkers launches goroutines for email processing and notification cleanup.
func startWorkers(ctx context.Context, db *sql.DB, provider MailProvider) {
	log.Printf("Starting email worker")
	safeGo(func() { emailQueueWorker(ctx, New(db), provider, time.Minute) })
	log.Printf("Starting notification purger worker")
	safeGo(func() { notificationPurgeWorker(ctx, New(db), time.Hour) })
}
