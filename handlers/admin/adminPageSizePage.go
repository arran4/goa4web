package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/config"
)

// AdminPageSizePage allows administrators to adjust pagination limits.
// The change only affects the in-memory configuration. Update the
// configuration file separately to persist the values.
func AdminPageSizePage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		min, _ := strconv.Atoi(r.PostFormValue("min"))
		max, _ := strconv.Atoi(r.PostFormValue("max"))
		def, _ := strconv.Atoi(r.PostFormValue("default"))
		config.UpdatePaginationConfig(&config.AppRuntimeConfig, min, max, def)

		data := struct {
			*common.CoreData
			Messages []string
			Back     string
		}{
			CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
			Back:     "/admin/page-size",
			Messages: []string{"Pagination settings updated in memory. Update the configuration file to persist."},
		}
		handlers.TemplateHandler(w, r, "admin/runTaskPage.gohtml", data)
		return
	}

	data := struct {
		*common.CoreData
		Min     int
		Max     int
		Default int
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Min:      config.AppRuntimeConfig.PageSizeMin,
		Max:      config.AppRuntimeConfig.PageSizeMax,
		Default:  config.AppRuntimeConfig.PageSizeDefault,
	}
	handlers.TemplateHandler(w, r, "admin/pageSizePage.gohtml", data)
}
