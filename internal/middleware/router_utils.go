package middleware

import "net/http"

// RouterWrapper wraps a router with additional middleware.
type RouterWrapper interface {
	Wrap(http.Handler) http.Handler
}

// RouterWrapperFunc allows ordinary functions to satisfy RouterWrapper.
type RouterWrapperFunc func(http.Handler) http.Handler

// Wrap executes the underlying middleware function.
func (f RouterWrapperFunc) Wrap(h http.Handler) http.Handler { return f(h) }

// NewMiddlewareChain returns a RouterWrapper that wraps a handler with the provided
// middleware functions in the order supplied.
func NewMiddlewareChain(mw ...func(http.Handler) http.Handler) RouterWrapper {
	return RouterWrapperFunc(func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	})
}
