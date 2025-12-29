package statsworker

import (
	"context"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	// Just ensure it doesn't panic and exits when context is done
	Worker(ctx)
}
