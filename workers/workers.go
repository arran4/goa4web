package workers

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/scheduler"

	"github.com/arran4/goa4web/workers/auditworker"
	"github.com/arran4/goa4web/workers/backgroundtaskworker"
	"github.com/arran4/goa4web/workers/emailqueue"
	"github.com/arran4/goa4web/workers/externallinkworker"
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
func Start(ctx context.Context, q db.Querier, provider email.Provider, dlqProvider dlq.DLQ, cfg *config.RuntimeConfig, bus *eventbus.Bus) {
	log.Printf("Starting email worker")
	safeGo(func() {
		emailqueue.StartEventListener(ctx, q, provider, dlqProvider, bus, cfg)
	})
	log.Printf("Starting generic scheduler and digest consumer")
	safeGo(func() {
		var nOpts []notifications.Option
		nOpts = append(nOpts, notifications.WithQueries(q))
		if cq, ok := q.(db.CustomQueries); ok {
			nOpts = append(nOpts, notifications.WithCustomQueries(cq))
		}
		nOpts = append(nOpts, notifications.WithEmailProvider(provider), notifications.WithBus(bus), notifications.WithConfig(cfg))
		n := notifications.New(nOpts...)
		// Start consumer
		consumer := notifications.NewDigestConsumer(n)
		go consumer.Run(ctx)

		// Start generic scheduler
		s := scheduler.New(q)
		s.Register(scheduler.Task{
			Name:    notifications.SchedulerTaskName,
			Handler: n.ScheduleDigest,
			Type:    scheduler.TaskTypeBackfill,
		})
		s.Register(scheduler.Task{
			Name:     "notification_purge",
			Handler:  n.PurgeReadNotifications,
			Type:     scheduler.TaskTypePeriodic,
			Interval: time.Hour,
		})
		s.Register(scheduler.Task{
			Name: "email_queue_poll",
			Handler: func(ctx context.Context, t time.Time) error {
				emailqueue.ProcessPendingEmail(ctx, q, provider, dlqProvider, cfg)
				return nil
			},
			Type:      scheduler.TaskTypePeriodic,
			Interval:  time.Duration(cfg.EmailWorkerInterval) * time.Second,
			Ephemeral: true,
		})
		s.Run(ctx, 1*time.Second)
	})
	log.Printf("Starting event bus logger worker")
	safeGo(func() { logworker.Worker(ctx, bus) })
	log.Printf("Starting audit worker")
	safeGo(func() { auditworker.Worker(ctx, bus, q) })
	log.Printf("Starting notification bus worker")
	safeGo(func() {
		var nOpts []notifications.Option
		nOpts = append(nOpts, notifications.WithQueries(q))
		if cq, ok := q.(db.CustomQueries); ok {
			nOpts = append(nOpts, notifications.WithCustomQueries(cq))
		}
		nOpts = append(nOpts, notifications.WithEmailProvider(provider), notifications.WithBus(bus), notifications.WithConfig(cfg))
		n := notifications.New(nOpts...)
		n.BusWorker(ctx, bus, dlqProvider)
	})
	log.Printf("Starting search index worker")
	safeGo(func() { searchworker.Worker(ctx, bus, q) })
	log.Printf("Starting background task worker")
	safeGo(func() { backgroundtaskworker.Worker(ctx, bus, q) })
	log.Printf("Starting post count worker")
	safeGo(func() { postcountworker.Worker(ctx, bus, q) })
	log.Printf("Starting external link worker")
	safeGo(func() { externallinkworker.Worker(ctx, bus, q, cfg) })
}
