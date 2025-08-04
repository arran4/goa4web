package user

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// userPublicProfilePage shows a public user profile.
func userPublicProfilePage(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !u.PublicProfileEnabledAt.Valid {
		http.NotFound(w, r)
		return
	}
	if _, err := queries.GetPublicProfileRoleForUser(r.Context(), u.Idusers); err != nil {
		http.NotFound(w, r)
		return
	}
	cd.PageTitle = fmt.Sprintf("Profile for %s", u.Username.String)
	data := struct {
		User *db.User
	}{
		User: &db.User{Idusers: u.Idusers, Username: u.Username},
	}
	handlers.TemplateHandler(w, r, "user/publicProfile.gohtml", data)
}
