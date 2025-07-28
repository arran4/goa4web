package admin

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// adminUserWritingsPage lists all writings authored by a user.
func adminUserWritingsPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	user, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	rows, err := queries.GetAllWritingsByUser(r.Context(), db.GetAllWritingsByUserParams{
		ViewerID:      int32(id),
		AuthorID:      int32(id),
		ViewerMatchID: sql.NullInt32{Int32: int32(id), Valid: true},
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		User     *db.User
		Writings []*db.GetAllWritingsByUserRow
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Writings: rows,
	}
	handlers.TemplateHandler(w, r, "userWritingsPage.gohtml", data)
}
