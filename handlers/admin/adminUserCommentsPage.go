package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserCommentsPage lists all comments posted by a user.
func adminUserCommentsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cpu := cd.CurrentProfileUser()
	if cpu.Idusers == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	queries := cd.Queries()
	rows, err := queries.AdminGetAllCommentsByUser(r.Context(), cpu.Idusers)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Comments by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User     *db.User
		Comments []*db.AdminGetAllCommentsByUserRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Comments: rows,
	}
	handlers.TemplateHandler(w, r, "userCommentsPage.gohtml", data)
}
