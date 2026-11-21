package forum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func setPublicForumFollow(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.UserID == 0 || !cd.CanLogIn() {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if !cd.Config.NotificationsEnabled {
		http.NotFound(w, r)
		return
	}
	follow := r.PostFormValue("follow")
	if follow != "yes" && follow != "no" {
		http.Error(w, "Choose Follow or Unfollow.", http.StatusBadRequest)
		return
	}
	if err := cd.SetPublicForumNewThreadFollow(follow == "yes"); err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}
