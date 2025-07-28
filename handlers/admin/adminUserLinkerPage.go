package admin

import (
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
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	user, err := queries.GetUserById(r.Context(), int32(id))
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
	data := struct {
		*common.CoreData
		User  *db.User
		Links []*db.GetLinkerItemsByUserDescendingRow
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Links:    rows,
	}
	handlers.TemplateHandler(w, r, "userLinkerPage.gohtml", data)
}
