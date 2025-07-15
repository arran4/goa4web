package user

import (
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/handlers/common"

	"github.com/arran4/goa4web/config"
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
	common.TemplateHandler(w, r, "pageSizePage.gohtml", data)
}

func userPageSizeSaveActionPage(w http.ResponseWriter, r *http.Request) {
	userPageSizePage(w, r)
}
