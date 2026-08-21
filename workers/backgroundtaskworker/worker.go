package backgroundtaskworker

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// DefaultConcurrency is the default maximum number of concurrently running background tasks.
const DefaultConcurrency = 8

// DefaultBufferSize is the default buffer size for the background task queue.
const DefaultBufferSize = 100

// Option configures the background task worker.
type Option func(*WorkerConfig)

// WorkerConfig holds configuration for the background task worker.
type WorkerConfig struct {
	HTTPClient  *http.Client
	CoreOptions []common.CoreOption
	Concurrency int
	BufferSize  int
	Ready       chan<- struct{}
}

// WithReady configures a channel that is closed once the worker has subscribed and is ready.
func WithReady(ready chan<- struct{}) Option {
	return func(c *WorkerConfig) {
		c.Ready = ready
	}
}

// WithHTTPClient sets the HTTP client to provide to CoreData during background task execution.
func WithHTTPClient(client *http.Client) Option {
	return func(c *WorkerConfig) {
		c.HTTPClient = client
	}
}

// WithConcurrency sets the maximum concurrency level for background task execution.
func WithConcurrency(n int) Option {
	return func(c *WorkerConfig) {
		if n > 0 {
			c.Concurrency = n
		}
	}
}

// WithBufferSize sets the reliable queue buffer size for the background task worker.
func WithBufferSize(n int) Option {
	return func(c *WorkerConfig) {
		if n > 0 {
			c.BufferSize = n
		}
	}
}

// WithCoreOptions supplies additional CoreData options for background task execution.
func WithCoreOptions(opts ...common.CoreOption) Option {
	return func(c *WorkerConfig) {
		c.CoreOptions = append(c.CoreOptions, opts...)
	}
}

// Worker listens for task events implementing tasks.BackgroundTasker.
// The background method is executed with bounded concurrency and any returned task
// is published back onto the bus when the work completes.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier, cfg *config.RuntimeConfig, opts ...Option) {
	if bus == nil || q == nil {
		return
	}
	wc := &WorkerConfig{
		Concurrency: DefaultConcurrency,
		BufferSize:  DefaultBufferSize,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(wc)
		}
	}

	sub := bus.SubscribeWithOptions(
		eventbus.WithTypes(eventbus.TaskMessageType),
		eventbus.WithReliableDelivery(),
		eventbus.WithBufferSize(wc.BufferSize),
		eventbus.WithFilter(func(msg eventbus.Message) bool {
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				return false
			}
			_, ok = evt.Task.(tasks.BackgroundTasker)
			return ok
		}),
	)
	defer sub.Close()

	if wc.Ready != nil {
		close(wc.Ready)
	}

	ch := sub.Channel()
	sem := make(chan struct{}, wc.Concurrency)
	var wg sync.WaitGroup

	for {
		select {
		case env, ok := <-ch:
			if !ok {
				wg.Wait()
				return
			}
			evt, ok := env.Msg.(eventbus.TaskEvent)
			if !ok {
				env.Ack()
				continue
			}
			p, ok := evt.Task.(tasks.BackgroundTasker)
			if !ok {
				env.Ack()
				continue
			}

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				env.Ack()
				wg.Wait()
				return
			}

			wg.Add(1)
			go func(env eventbus.Envelope, evt eventbus.TaskEvent, p tasks.BackgroundTasker) {
				defer func() {
					<-sem
					wg.Done()
					env.Ack()
				}()

				evtCtx := context.WithoutCancel(ctx)
				coreOpts := append([]common.CoreOption(nil), wc.CoreOptions...)
				if wc.HTTPClient != nil {
					coreOpts = append(coreOpts, common.WithHTTPClient(wc.HTTPClient))
				}
				cd := common.NewCoreData(evtCtx, q, cfg, coreOpts...)
				evtCtx = context.WithValue(evtCtx, consts.KeyCoreData, cd)

				t, err := p.BackgroundTask(evtCtx, q)
				if err != nil {
					log.Printf("background task: %v", err)
					return
				}
				if t != nil {
					nEvt := eventbus.TaskEvent{
						Path:    evt.Path,
						Task:    t,
						UserID:  evt.UserID,
						Time:    time.Now(),
						Data:    evt.Data,
						Outcome: eventbus.TaskOutcomeSuccess,
					}
					if err := bus.Publish(nEvt); err != nil && err != eventbus.ErrBusClosed {
						log.Printf("background publish: %v", err)
					}
				}
			}(env, evt, p)

		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}
