package main

import (
	"log"
	"net/http"
	"strconv"
)

func userPageSizePage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		*CoreData
		Min int
		Max int
		Def int
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Min:      appPaginationConfig.Min,
		Max:      appPaginationConfig.Max,
		Def:      appPaginationConfig.Default,
	}
	if err := renderTemplate(w, r, "userPageSizePage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userPageSizeSaveActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/page-size", http.StatusSeeOther)
		return
	}
	min, _ := strconv.Atoi(r.FormValue("min"))
	max, _ := strconv.Atoi(r.FormValue("max"))
	def, _ := strconv.Atoi(r.FormValue("default"))
	cfg := resolvePaginationConfig(PaginationConfig{Min: min, Max: max, Default: def}, PaginationConfig{}, PaginationConfig{})
	appPaginationConfig = cfg
	http.Redirect(w, r, "/usr/page-size", http.StatusSeeOther)
}
