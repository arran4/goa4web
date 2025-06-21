package main

import (
	"log"
	"net/http"
)

// renderTemplate executes the named template with data and writes the result to w.
// Any execution error is logged and a 500 response is sent to the client.
func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
