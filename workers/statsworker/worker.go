package statsworker

import (
	"context"
	"time"

	"github.com/arran4/goa4web/internal/stats"
)

// Worker dumps stats periodically.
func Worker(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			stats.Dump()
			return
		case <-ticker.C:
			stats.Dump()
		}
	}
}
