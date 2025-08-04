package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminUserSubscriptionsPage lists subscription patterns of a user.
func adminUserSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	uid := cd.CurrentProfileUserID()
	if uid == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	user := cd.CurrentProfileUser()
	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	queries := cd.Queries()
	rows, err := queries.ListSubscriptionsByUser(r.Context(), uid)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cd.PageTitle = fmt.Sprintf("Subscriptions of %s", user.Username.String)
	data := struct {
		*common.CoreData
		User *db.User
		Subs []*db.ListSubscriptionsByUserRow
	}{
		CoreData: cd,
		User:     &db.User{Idusers: uid, Username: user.Username},
		Subs:     rows,
	}
	handlers.TemplateHandler(w, r, "userSubscriptionsPage.gohtml", data)
}
