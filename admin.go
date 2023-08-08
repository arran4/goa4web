package main

import (
	"golang.org/x/exp/slices"
	"net/http"
)

func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cd := request.Context().Value(ContextValues("coreData")).(*CoreData)
		levelRequired := []string{"administrator"}
		if !slices.Contains(levelRequired, cd.SecurityLevel) {
			http.Error(writer, "Incorrect security level", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request.WithContext(request.Context()))
	})
}
