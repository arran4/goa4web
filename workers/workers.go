package workers

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
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

// Option configures background workers started by Start.
type Option func(*WorkersConfig)

// WorkersConfig holds worker configuration.
type WorkersConfig struct {
	HTTPClient  *http.Client
	CoreOptions []common.CoreOption
}

// Group owns worker and auxiliary goroutines so shutdown can cancel and bound their wait.
type Group struct {
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	stopping bool
	wg       sync.WaitGroup
}

// NewGroup creates a managed goroutine group whose cancellation is explicit.
func NewGroup(parent context.Context) *Group {
	ctx, cancel := context.WithCancel(parent)
	return &Group{ctx: ctx, cancel: cancel}
}

// Go starts shutdown-tracked work using the worker lifetime context.
func (g *Group) Go(fn func(context.Context)) bool {
	if g == nil || fn == nil {
		return false
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.stopping {
		return false
	}
	g.wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panic: %v", r)
				os.Exit(1)
			}
		}()
		fn(g.ctx)
	})
	return true
}

// Cancel prevents new tracked work and cooperatively cancels running work.
func (g *Group) Cancel() {
	if g == nil {
		return
	}
	g.mu.Lock()
	if !g.stopping {
		g.stopping = true
		g.cancel()
	}
	g.mu.Unlock()
}

// Wait waits for all tracked work or for ctx to expire.
func (g *Group) Wait(ctx context.Context) error {
	if g == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WithHTTPClient sets the HTTP client to supply to workers making external requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *WorkersConfig) {
		c.HTTPClient = client
	}
}

// WithCoreOptions supplies additional CoreData options for background workers.
func WithCoreOptions(opts ...common.CoreOption) Option {
	return func(c *WorkersConfig) {
		c.CoreOptions = append(c.CoreOptions, opts...)
	}
}

// Start launches all background workers using the given configuration.
func Start(ctx context.Context, q db.Querier, provider email.Provider, dlqProvider dlq.DLQ, cfg *config.RuntimeConfig, bus *eventbus.Bus, opts ...Option) *Group {
	wc := &WorkersConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(wc)
		}
	}
	group := NewGroup(ctx)

	log.Printf("Starting email worker")
	group.Go(func(ctx context.Context) {
		emailqueue.StartEventListener(ctx, q, provider, dlqProvider, bus, cfg)
	})
	log.Printf("Starting generic scheduler and digest consumer")
	group.Go(func(ctx context.Context) {
		n := notifications.New(
			notifications.WithQueries(q),
			notifications.WithEmailProvider(provider),
			notifications.WithBus(bus),
			notifications.WithConfig(cfg),
		)
		// Start consumer
		consumer := notifications.NewDigestConsumer(n)
		group.Go(func(ctx context.Context) { consumer.Run(ctx) })

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
		s.Register(scheduler.Task{
			Name: "image_cache_fetch",
			Handler: func(ctx context.Context, t time.Time) error {
				cd := common.NewCoreData(ctx, q, cfg)
				return cd.ProcessPendingRemoteImageCacheEntries(ctx, 10)
			},
			Type:      scheduler.TaskTypePeriodic,
			Interval:  time.Minute,
			Ephemeral: true,
		})
		s.Run(ctx, 1*time.Second)
	})
	log.Printf("Starting event bus logger worker")
	group.Go(func(ctx context.Context) { logworker.Worker(ctx, bus) })
	log.Printf("Starting audit worker")
	group.Go(func(ctx context.Context) { auditworker.Worker(ctx, bus, q) })
	log.Printf("Starting notification bus worker")
	group.Go(func(ctx context.Context) {
		n := notifications.New(
			notifications.WithQueries(q),
			notifications.WithEmailProvider(provider),
			notifications.WithBus(bus),
			notifications.WithConfig(cfg),
		)
		n.BusWorker(ctx, bus, dlqProvider)
	})
	log.Printf("Starting search index worker")
	group.Go(func(ctx context.Context) { searchworker.Worker(ctx, bus, q) })
	log.Printf("Starting background task worker")
	group.Go(func(ctx context.Context) {
		var bopts []backgroundtaskworker.Option
		if wc.HTTPClient != nil {
			bopts = append(bopts, backgroundtaskworker.WithHTTPClient(wc.HTTPClient))
		}
		if len(wc.CoreOptions) > 0 {
			bopts = append(bopts, backgroundtaskworker.WithCoreOptions(wc.CoreOptions...))
		}
		backgroundtaskworker.Worker(ctx, bus, q, cfg, bopts...)
	})
	log.Printf("Starting post count worker")
	group.Go(func(ctx context.Context) { postcountworker.Worker(ctx, bus, q) })
	log.Printf("Starting external link worker")
	group.Go(func(ctx context.Context) { externallinkworker.Worker(ctx, bus, q, cfg) })
	return group
}
