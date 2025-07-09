package admin

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func adminUserProfilePage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	user, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	emails, _ := queries.GetUserEmailsByUserID(r.Context(), int32(id))
	data := struct {
		*corecommon.CoreData
		User   *db.User
		Emails []*db.UserEmail
	}{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		User:     &db.User{Idusers: user.Idusers, Username: user.Username},
		Emails:   emails,
	}
	if err := templates.RenderTemplate(w, "admin/userProfile.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
