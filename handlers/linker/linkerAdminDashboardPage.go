package linker

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// AdminDashboardPage shows summary links for linker admin.
func AdminDashboardPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Linker Admin"
	data := struct {
		CategoryCount int
		LinkCount     int
	}{}
	if rows, err := cd.LinkerCategoryCounts(); err == nil {
		data.CategoryCount = len(rows)
		for _, c := range rows {
			data.LinkCount += int(c.Linkcount)
		}
	}
	handlers.TemplateHandler(w, r, "linkerAdminDashboardPage.gohtml", data)
}
