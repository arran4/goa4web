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
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	queries := cd.Queries()
	rows, err := queries.AdminGetAllBlogEntriesByUser(r.Context(), user.Idusers)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Blogs by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User  *db.User
		Blogs []*db.AdminGetAllBlogEntriesByUserRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Blogs:    rows,
	}
	handlers.TemplateHandler(w, r, "userBlogsPage.gohtml", data)
}
