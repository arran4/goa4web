package admin

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/config"
)

// AdminPageSizePage allows administrators to adjust pagination limits.
// The change only affects the in-memory configuration. Update the
// configuration file separately to persist the values.
func AdminPageSizePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Page Size"
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
			return
		}
		min, _ := strconv.Atoi(r.PostFormValue("min"))
		max, _ := strconv.Atoi(r.PostFormValue("max"))
		def, _ := strconv.Atoi(r.PostFormValue("default"))
		config.UpdatePaginationConfig(cd.Config, min, max, def)

		data := struct {
			Errors   []string
			Messages []string
			Back     string
		}{
			Back:     "/admin/page-size",
			Messages: []string{"Pagination settings updated in memory. Update the configuration file to persist."},
		}
		RunTaskPageTmpl.Handle(w, r, data)
		return
	}

	data := struct {
		Min     int
		Max     int
		Default int
	}{
		Min:     cd.Config.PageSizeMin,
		Max:     cd.Config.PageSizeMax,
		Default: cd.Config.PageSizeDefault,
	}
	AdminPageSizePageTmpl.Handle(w, r, data)
}

const AdminPageSizePageTmpl handlers.Page = "admin/pageSizePage.gohtml"

const RunTaskPageTmpl handlers.Page = "admin/runTaskPage.gohtml"
