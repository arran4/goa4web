package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserBlogsPage lists all blog posts authored by a user.
func adminUserBlogsPage(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.AdminGetAllBlogEntriesByUser(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Blogs by %s", user.Username.String)
	data := struct {
		User  *db.User
		Blogs []*db.AdminGetAllBlogEntriesByUserRow
	}{
		User:  &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Blogs: rows,
	}
	AdminUserBlogsPageTmpl.Handle(w, r, data)
}

const AdminUserBlogsPageTmpl handlers.Page = "admin/userBlogsPage.gohtml"
