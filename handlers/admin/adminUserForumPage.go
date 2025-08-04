package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserForumPage lists forum threads started by a user.
func adminUserForumPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	uid := cd.CurrentProfileUserID()
	if uid == 0 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	queries := cd.Queries()
	rows, err := queries.AdminGetThreadsStartedByUserWithTopic(r.Context(), uid)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Forum threads by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User    *db.User
		Threads []*db.AdminGetThreadsStartedByUserWithTopicRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: uid, Username: user.Username},
		Threads:  rows,
	}
	handlers.TemplateHandler(w, r, "userForumPage.gohtml", data)
}
