package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminTemplateExportPage displays the template export form.
func AdminTemplateExportPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Template Export"

	data := struct {
		SelectedSet    string
		SelectedFormat string
	}{
		SelectedSet:    "site",
		SelectedFormat: "zip",
	}

	AdminTemplateExportPageTmpl.Handle(w, r, data)
}

// AdminTemplateExportPageTmpl renders the template export page.
const AdminTemplateExportPageTmpl tasks.Template = "admin/templateExportPage.gohtml"
