package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// adminUserBlogsPage lists all blog posts authored by a user.
func adminUserBlogsPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	user, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	rows, err := queries.GetAllBlogEntriesByUserForAdmin(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Blogs by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User  *db.User
		Blogs []*db.GetAllBlogEntriesByUserForAdminRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Blogs:    rows,
	}
	handlers.TemplateHandler(w, r, "userBlogsPage.gohtml", data)
}
