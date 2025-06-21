package main

import (
	"log"
	"net/http"
)

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid int32
		if u, ok := r.Context().Value(ContextValues("user")).(*User); ok && u != nil {
			uid = u.Idusers
		}
		log.Printf("%s %s uid=%d", r.Method, r.URL.Path, uid)
		next.ServeHTTP(w, r)
	})
}
