package main

import "net/http"

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) error {
	return getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, name, data)
}
