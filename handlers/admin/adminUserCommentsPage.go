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
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	queries := cd.Queries()
	rows, err := queries.AdminGetAllCommentsByUser(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Comments by %s", user.Username.String)
	data := struct {
		User     *db.User
		Comments []*db.AdminGetAllCommentsByUserRow
	}{
		User:     &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Comments: rows,
	}
	handlers.TemplateHandler(w, r, "userCommentsPage.gohtml", data)
}
