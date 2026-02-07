package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminTemplateExportPage struct{}

func (p *AdminTemplateExportPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Template Export"

	data := struct {
		SelectedSet    string
		SelectedFormat string
	}{
		SelectedSet:    "site",
		SelectedFormat: "zip",
	}

	AdminTemplateExportPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminTemplateExportPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Template Export", "/admin/templates/export", &AdminEmailTemplatePage{}
}

func (p *AdminTemplateExportPage) PageTitle() string {
	return "Template Export"
}

var _ common.Page = (*AdminTemplateExportPage)(nil)
var _ http.Handler = (*AdminTemplateExportPage)(nil)

// AdminTemplateExportPageTmpl renders the template export page.
const AdminTemplateExportPageTmpl tasks.Template = "admin/templateExportPage.gohtml"
