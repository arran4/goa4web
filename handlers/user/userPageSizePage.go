package user

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
)

func userPageSizePage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			min, _ := strconv.Atoi(r.PostFormValue("min"))
			max, _ := strconv.Atoi(r.PostFormValue("max"))
			def, _ := strconv.Atoi(r.PostFormValue("default"))
			config.UpdatePaginationConfig(&config.AppRuntimeConfig, min, max, def)
		}
		http.Redirect(w, r, "/usr/page-size", http.StatusSeeOther)
		return
	}
	data := struct {
		*common.CoreData
		Min     int
		Max     int
		Default int
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Min:      config.AppRuntimeConfig.PageSizeMin,
		Max:      config.AppRuntimeConfig.PageSizeMax,
		Default:  config.AppRuntimeConfig.PageSizeDefault,
	}
	if err := templates.RenderTemplate(w, "pageSizePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func userPageSizeSaveActionPage(w http.ResponseWriter, r *http.Request) {
	userPageSizePage(w, r)
}
