package externallink

import (
	"time"

	"github.com/arran4/goa4web/internal/opengraph"
)

// FetchMetadata obtains Open Graph metadata using the established retry policy.
func FetchMetadata(url string, fetch func(string) (*opengraph.Info, error), pause func(time.Duration)) (*opengraph.Info, error) {
	var err error
	for attempt := range 3 {
		var info *opengraph.Info
		info, err = fetch(url)
		if err == nil {
			return info, nil
		}
		pause(time.Duration(attempt+1) * 2 * time.Second)
	}
	return nil, err
}
