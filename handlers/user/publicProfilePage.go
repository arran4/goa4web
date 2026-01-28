package user

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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
	isOwner := cd.UserID == u.Idusers
	var profileOff bool
	if !u.PublicProfileEnabledAt.Valid {
		if !isOwner {
			http.NotFound(w, r)
			return
		}
		profileOff = true
	}
	if _, err := queries.GetPublicProfileRoleForUser(r.Context(), u.Idusers); err != nil {
		if !isOwner {
			http.NotFound(w, r)
			return
		}
		profileOff = true
	}
	cd.PageTitle = fmt.Sprintf("Profile for %s", u.Username.String)
	data := struct {
		User       *db.User
		ProfileOff bool
	}{
		User:       &db.User{Idusers: u.Idusers, Username: u.Username},
		ProfileOff: profileOff,
	}
	PublicProfilePage.Handle(w, r, data)
}

const PublicProfilePage tasks.Template = "user/publicProfile.gohtml"
