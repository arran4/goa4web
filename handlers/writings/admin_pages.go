package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/admincommon"
)

// AdminWritingsPage renders the writings admin index with role summaries.
func AdminWritingsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CanPost   bool
		UserRoles []admincommon.UserRoleInfo
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writings Admin"
	data := Data{CanPost: cd.HasGrant("writing", "post", "edit", 0) && cd.AdminMode}

	queries := cd.Queries()
	userRoles, err := admincommon.LoadUserRoleInfo(r.Context(), queries, func(role string, isAdmin bool) bool {
		return isAdmin || role == "content writer"
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.UserRoles = userRoles
	sort.Slice(data.UserRoles, func(i, j int) bool {
		return data.UserRoles[i].Username.String < data.UserRoles[j].Username.String
	})

	handlers.TemplateHandler(w, r, WritingsAdminPageTmpl, data)
}
