package main

import (
	"golang.org/x/exp/slices"
	"log"
	"net/http"
)

func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cd := request.Context().Value(ContextValues("coreData")).(*CoreData)
		cd.AdminChecked = true
		levelRequired := []string{"administrator"}
		if !slices.Contains(levelRequired, cd.SecurityLevel) {
			err := compiledTemplates.ExecuteTemplate(writer, "adminNoAccessPage.tmpl", request.Context().Value(ContextValues("coreData")).(*CoreData))
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
