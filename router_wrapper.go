package goa4web

import "net/http"

// routerWrapper wraps a router with additional middleware.
type routerWrapper interface {
	Wrap(http.Handler) http.Handler
}

// routerWrapperFunc allows ordinary functions to satisfy routerWrapper.
type routerWrapperFunc func(http.Handler) http.Handler

func (f routerWrapperFunc) Wrap(h http.Handler) http.Handler { return f(h) }

// newMiddlewareChain returns a routerWrapper that wraps a handler with the provided
// middleware functions in the order supplied.
func newMiddlewareChain(mw ...func(http.Handler) http.Handler) routerWrapper {
	return routerWrapperFunc(func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	})
}
