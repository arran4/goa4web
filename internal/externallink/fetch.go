package externallink

import (
	"context"
	"time"

	"github.com/arran4/goa4web/internal/opengraph"
)

// FetchMetadata obtains Open Graph metadata using the established retry policy.
func FetchMetadata(ctx context.Context, url string, fetch func(context.Context, string) (*opengraph.Info, error), pause func(context.Context, time.Duration) error) (*opengraph.Info, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var err error
	for attempt := range 3 {
		var info *opengraph.Info
		info, err = fetch(ctx, url)
		if err == nil {
			return info, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if err := pause(ctx, time.Duration(attempt+1)*2*time.Second); err != nil {
			return nil, err
		}
	}
	return nil, err
}

// Wait pauses for d or returns as soon as ctx is cancelled.
func Wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
