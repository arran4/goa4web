package jobs

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/opengraph"
)

// maxFetchRetryAfter caps the delay requested by an upstream server.
const maxFetchRetryAfter = 5 * time.Minute

func fetchRetryPolicy(jobType string, err error, now time.Time) (permanent bool, retryAfter time.Duration) {
	if jobType != JobTypeExternalLink && jobType != JobTypeImageCache {
		return false, 0
	}
	if errors.Is(err, opengraph.ErrInvalidDestination) {
		return true, 0
	}
	var status *opengraph.HTTPStatusError
	if !errors.As(err, &status) {
		return false, 0
	}
	switch status.StatusCode {
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return false, parseFetchRetryAfter(status.RetryAfter, now)
	case http.StatusRequestTimeout, http.StatusTooEarly:
		return false, 0
	default:
		return status.StatusCode >= 300 && status.StatusCode < 500, 0
	}
}

func parseFetchRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	seconds, err := strconv.ParseUint(value, 10, 64)
	if err == nil || errors.Is(err, strconv.ErrRange) {
		return time.Duration(min(seconds, uint64(maxFetchRetryAfter/time.Second))) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil {
		return max(0, min(at.Sub(now), maxFetchRetryAfter))
	}
	return 0
}
