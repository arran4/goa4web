package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserImagebbsPage lists image board posts by a user.
func adminUserImagebbsPage(w http.ResponseWriter, r *http.Request) {
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
	rows, err := queries.GetImagePostsByUserDescendingAll(r.Context(), db.GetImagePostsByUserDescendingAllParams{
		UsersIdusers: uid,
		Limit:        100,
		Offset:       0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Images by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User  *db.User
		Posts []*db.GetImagePostsByUserDescendingAllRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: uid, Username: user.Username},
		Posts:    rows,
	}
	handlers.TemplateHandler(w, r, "userImagebbsPage.gohtml", data)
}
