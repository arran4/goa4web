package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserWritingsPage lists all writings authored by a user.
func adminUserWritingsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cpu := cd.CurrentProfileUser()
	if cpu.Idusers == 0 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	queries := cd.Queries()
	rows, err := queries.AdminGetAllWritingsByAuthor(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Writings by %s", user.Username.String)
	data := struct {
		User     *db.User
		Writings []*db.AdminGetAllWritingsByAuthorRow
	}{
		User:     &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Writings: rows,
	}
	AdminUserWritingsPageTmpl.Handle(w, r, data)
}

const AdminUserWritingsPageTmpl handlers.Page = "admin/userWritingsPage.gohtml"
