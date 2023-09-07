package main

import (
	"golang.org/x/exp/slices"
	"log"
	"net/http"
)

func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cd := request.Context().Value(ContextValues("coreData")).(*CoreData)
		levelRequired := []string{"administrator"}
		if !slices.Contains(levelRequired, cd.SecurityLevel) {
			err := getCompiledTemplates(NewFuncs(request)).ExecuteTemplate(writer, "adminNoAccessPage.gohtml", request.Context().Value(ContextValues("coreData")).(*CoreData))
			if err != nil {
				log.Printf("Template Error: %s", err)
				http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}
		next.ServeHTTP(writer, request.WithContext(request.Context()))
	})
}
