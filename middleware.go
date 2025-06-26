package goa4web

import (
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
)

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

// RequestLoggerMiddleware logs incoming requests and the associated user ID.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid int32
		if u, ok := r.Context().Value(common.KeyUser).(*User); ok && u != nil {
			uid = u.Idusers
		}
		log.Printf("%s %s uid=%d", r.Method, r.URL.Path, uid)
		next.ServeHTTP(w, r)
	})
}
