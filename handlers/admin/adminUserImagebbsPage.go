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
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	queries := cd.Queries()
	rows, err := queries.GetImagePostsByUserDescendingAll(r.Context(), db.GetImagePostsByUserDescendingAllParams{
		UsersIdusers: user.Idusers,
		Limit:        100,
		Offset:       0,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Images by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User  *db.User
		Posts []*db.GetImagePostsByUserDescendingAllRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Posts:    rows,
	}
	handlers.TemplateHandler(w, r, "userImagebbsPage.gohtml", data)
}
