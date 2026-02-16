package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"
)

// BenchmarkShutdown measures the latency of Shutdown when waiting for a consumer
// to drain the channel.
func BenchmarkShutdown(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bus := NewBus()
		ctx, cancel := context.WithCancel(context.Background())

		// Subscribe
		ch := bus.Subscribe(TaskMessageType)

		var wg sync.WaitGroup
		wg.Add(1)

		// Consumer that delays reading
		go func() {
			defer wg.Done()
			// Wait 5ms before starting to consume
			time.Sleep(5 * time.Millisecond)
			for {
				select {
				case <-ctx.Done():
					return
				case env, ok := <-ch:
					if !ok {
						return
					}
					// Consumed immediately
					env.Ack()
				}
			}
		}()

		// Publish a message to fill buffer (size 1)
		bus.Publish(TaskEvent{Task: nil})

		// Shutdown
		// Current: checks len=1. Sleeps 10ms. Checks len=0. Returns. Time ~10ms.
		// Optimized: Waits for signal. Consumer reads at 5ms. Signal at 5ms. Returns. Time ~5ms.
		if err := bus.Shutdown(context.Background()); err != nil {
			b.Fatalf("shutdown failed: %v", err)
		}

		cancel()
		wg.Wait()
	}
}
