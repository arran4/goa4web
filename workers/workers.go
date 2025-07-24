package workers

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	email "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"

	"github.com/arran4/goa4web/workers/auditworker"
	"github.com/arran4/goa4web/workers/emailqueue"
	"github.com/arran4/goa4web/workers/logworker"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

// safeGo runs fn in a goroutine and exits the program if the goroutine panics.
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panic: %v", r)
				os.Exit(1)
			}
		}()
		fn()
	}()
}

// Start launches all background workers using the given configuration.
func Start(ctx context.Context, db *sql.DB, provider email.Provider, dlqProvider dlq.DLQ, cfg config.RuntimeConfig, bus *eventbus.Bus) {
	log.Printf("Starting email worker")
	safeGo(func() {
		emailqueue.EmailQueueWorker(ctx, dbpkg.New(db), provider, dlqProvider, bus, time.Duration(cfg.EmailWorkerInterval)*time.Second)
	})
	log.Printf("Starting notification purger worker")
	safeGo(func() {
		n := notifications.New(
			notifications.WithQueries(dbpkg.New(db)),
			notifications.WithEmailProvider(provider),
			notifications.WithBus(bus),
		)
		n.NotificationPurgeWorker(ctx, time.Hour)
	})
	log.Printf("Starting event bus logger worker")
	safeGo(func() { logworker.Worker(ctx, bus) })
	log.Printf("Starting audit worker")
	safeGo(func() { auditworker.Worker(ctx, bus, dbpkg.New(db)) })
	log.Printf("Starting notification bus worker")
	safeGo(func() {
		n := notifications.New(
			notifications.WithQueries(dbpkg.New(db)),
			notifications.WithEmailProvider(provider),
			notifications.WithBus(bus),
		)
		n.BusWorker(ctx, bus, dlqProvider)
	})
	log.Printf("Starting search index worker")
	safeGo(func() { searchworker.Worker(ctx, bus, dbpkg.New(db)) })
	log.Printf("Starting post count worker")
	safeGo(func() { postcountworker.Worker(ctx, bus, dbpkg.New(db)) })
}
