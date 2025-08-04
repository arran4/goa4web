package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserLinkerPage lists linker posts created by a user.
func adminUserLinkerPage(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.GetLinkerItemsByUserDescending(r.Context(), db.GetLinkerItemsByUserDescendingParams{
		UsersIdusers: uid,
		Limit:        100,
		Offset:       0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Links by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User  *db.User
		Links []*db.GetLinkerItemsByUserDescendingRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: uid, Username: user.Username},
		Links:    rows,
	}
	handlers.TemplateHandler(w, r, "userLinkerPage.gohtml", data)
}
