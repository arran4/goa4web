package linker

import (
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminDashboardPage shows summary links for linker admin.
func AdminDashboardPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Linker Admin"
	data := struct {
		CategoryCount int
		LinkCount     int
		Roles         []*db.Role
	}{}
	if rows, err := cd.LinkerCategoryCounts(); err == nil {
		data.CategoryCount = len(rows)
		for _, c := range rows {
			data.LinkCount += int(c.Linkcount)
		}
	}
	if roles, err := cd.AllRoles(); err == nil {
		for _, role := range roles {
			if strings.Contains(strings.ToLower(role.Name), "linker") {
				data.Roles = append(data.Roles, role)
			}
		}
	}
	LinkerAdminDashboardPageTmpl.Handle(w, r, data)
}

const LinkerAdminDashboardPageTmpl handlers.Page = "linker/linkerAdminDashboardPage.gohtml"
