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
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	queries := cd.Queries()
	rows, err := queries.GetLinkerItemsByUserDescending(r.Context(), db.GetLinkerItemsByUserDescendingParams{
		UsersIdusers: cd.CurrentProfileUserID(),
		Limit:        100,
		Offset:       0,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Links by %s", user.Username.String)
	data := struct {
		*common.CoreData
		User  *db.User
		Links []*db.GetLinkerItemsByUserDescendingRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Links:    rows,
	}
	handlers.TemplateHandler(w, r, "userLinkerPage.gohtml", data)
}
