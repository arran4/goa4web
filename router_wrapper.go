package main

import "net/http"

// routerWrapper wraps a router with additional middleware.
type routerWrapper interface {
	Wrap(http.Handler) http.Handler
}

// routerWrapperFunc allows ordinary functions to satisfy routerWrapper.
type routerWrapperFunc func(http.Handler) http.Handler

func (f routerWrapperFunc) Wrap(h http.Handler) http.Handler { return f(h) }
