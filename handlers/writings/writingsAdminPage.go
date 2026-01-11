package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/admincommon"
)

// AdminPage renders the writings administration page.
func AdminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Users []admincommon.UserRoleInfo
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writings Admin"
	data := Data{}

	queries := cd.Queries()
	userRoles, err := admincommon.LoadUserRoleInfo(r.Context(), queries, nil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("LoadUserRoleInfo Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data.Users = userRoles
	sort.Slice(data.Users, func(i, j int) bool {
		return data.Users[i].Username.String < data.Users[j].Username.String
	})

	WritingsAdminPageTmpl.Handle(w, r, data)
}

const WritingsAdminPageTmpl handlers.Page = "writings/writingsAdminPage.gohtml"
