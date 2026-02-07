package admin

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminPageSizePage struct{}

func (p *AdminPageSizePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		RunTaskPageTmpl.Handler(data).ServeHTTP(w, r)
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
	AdminPageSizePageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminPageSizePage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Pagination", "/admin/page-size", &AdminSiteSettingsPage{} // Assuming pagination is under settings
}

func (p *AdminPageSizePage) PageTitle() string {
	return "Page Size"
}

var _ common.Page = (*AdminPageSizePage)(nil)
var _ http.Handler = (*AdminPageSizePage)(nil)

const AdminPageSizePageTmpl tasks.Template = "admin/pageSizePage.gohtml"

const RunTaskPageTmpl tasks.Template = "admin/runTaskPage.gohtml"
