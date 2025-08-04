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

// adminUserLinkerPage lists linker posts created by a user.
func adminUserLinkerPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["user"]
	id, _ := strconv.Atoi(idStr)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	user, err := queries.SystemGetUserByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	rows, err := queries.GetLinkerItemsByUserDescending(r.Context(), db.GetLinkerItemsByUserDescendingParams{
		UsersIdusers: int32(id),
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
