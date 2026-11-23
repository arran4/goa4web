package opengraph

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrInvalidDestination identifies a request URL that cannot be fetched safely.
var ErrInvalidDestination = errors.New("invalid fetch destination")

// HTTPStatusError describes an unsuccessful response without its URL or body.
type HTTPStatusError struct {
	StatusCode int
	RetryAfter string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("upstream returned HTTP %d", e.StatusCode)
}

// CheckHTTPStatus rejects unsuccessful responses before their bodies are parsed.
func CheckHTTPStatus(resp *http.Response) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	return &HTTPStatusError{StatusCode: resp.StatusCode, RetryAfter: resp.Header.Get("Retry-After")}
}
