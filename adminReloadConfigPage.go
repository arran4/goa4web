package main

import (
	"log"
	"net/http"
)

func adminReloadConfigPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin",
	}

	cfgMap := loadAppConfigFile(configFile)
	srv.Config = loadRuntimeConfig(cfgMap)

	data.Messages = append(data.Messages, "Configuration reloaded")

	if err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
