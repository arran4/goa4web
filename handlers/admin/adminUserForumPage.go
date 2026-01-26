package admin

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserForumPage lists forum threads started by a user.
func adminUserForumPage(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.AdminGetThreadsStartedByUserWithTopic(r.Context(), cpu.Idusers)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum threads by %s", user.Username.String)
	data := struct {
		User    *db.User
		Threads []*db.AdminGetThreadsStartedByUserWithTopicRow
	}{
		User:    &db.User{Idusers: cpu.Idusers, Username: user.Username},
		Threads: rows,
	}
	AdminUserForumPageTmpl.Handle(w, r, data)
}

const AdminUserForumPageTmpl tasks.Template = "admin/userForumPage.gohtml"
