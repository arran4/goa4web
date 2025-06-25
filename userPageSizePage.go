package goa4web

import (
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/runtimeconfig"
)

func userPageSizePage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			min, _ := strconv.Atoi(r.PostFormValue("min"))
			max, _ := strconv.Atoi(r.PostFormValue("max"))
			def, _ := strconv.Atoi(r.PostFormValue("default"))
			runtimeconfig.UpdatePaginationConfig(&runtimeconfig.AppRuntimeConfig, min, max, def)
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
		Min:      runtimeconfig.AppRuntimeConfig.PageSizeMin,
		Max:      runtimeconfig.AppRuntimeConfig.PageSizeMax,
		Default:  runtimeconfig.AppRuntimeConfig.PageSizeDefault,
	}
	if err := templates.RenderTemplate(w, "pageSizePage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userPageSizeSaveActionPage(w http.ResponseWriter, r *http.Request) {
	userPageSizePage(w, r)
}
