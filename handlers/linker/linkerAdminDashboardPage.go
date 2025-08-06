package linker

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// AdminDashboardPage shows an overview with quick links.
func AdminDashboardPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Linker Dashboard"

	rows, _ := cd.LinkerCategoryCounts()
	var linkCount int
	for _, row := range rows {
		linkCount += int(row.Linkcount)
	}

	data := struct {
		CategoryCount int
		LinkCount     int
	}{
		CategoryCount: len(rows),
		LinkCount:     linkCount,
	}

	handlers.TemplateHandler(w, r, "linkerAdminDashboardPage.gohtml", data)
}
