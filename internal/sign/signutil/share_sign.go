package signutil

import (
	"fmt"
	"time"

	"github.com/arran4/goa4web/internal/sign"
)

// SignSharePath signs a share path and appends authentication parameters.
func SignSharePath(path string, key string, duration string, noExpiry bool) (string, error) {
	var opts []sign.SignOption
	if !noExpiry {
		if duration == "" {
			return "", fmt.Errorf("duration required")
		}
		d, err := time.ParseDuration(duration)
		if err != nil {
			return "", fmt.Errorf("parse duration: %w", err)
		}
		opts = append(opts, sign.WithExpiry(time.Now().Add(d)))
	}
	signed, err := SignAndAddPath(path, path, key, opts...)
	if err != nil {
		return "", fmt.Errorf("sign url: %w", err)
	}
	return signed, nil
}
