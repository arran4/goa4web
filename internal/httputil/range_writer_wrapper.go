package httputil

import (
	"net/http"
)

// WithRangeSupport wraps an http.HandlerFunc to provide Range request support.
func WithRangeSupport(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := NewRangeResponseWriter(w, r)
		h(rw, r)
		rw.Serve()
	}
}
