package goa4web

import (
	"log"
	"net/http"
	"strconv"
)

func userPageSizePage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			min, _ := strconv.Atoi(r.PostFormValue("min"))
			max, _ := strconv.Atoi(r.PostFormValue("max"))
			def, _ := strconv.Atoi(r.PostFormValue("default"))
			updatePaginationConfig(&appRuntimeConfig, min, max, def)
		}
		http.Redirect(w, r, "/usr/page-size", http.StatusSeeOther)
		return
	}
	data := struct {
		*CoreData
		Min     int
		Max     int
		Default int
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Min:      appRuntimeConfig.PageSizeMin,
		Max:      appRuntimeConfig.PageSizeMax,
		Default:  appRuntimeConfig.PageSizeDefault,
	}
	if err := renderTemplate(w, r, "pageSizePage.gohtml", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userPageSizeSaveActionPage(w http.ResponseWriter, r *http.Request) {
	userPageSizePage(w, r)
}
