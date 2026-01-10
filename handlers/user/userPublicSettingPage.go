package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// userPublicProfileSettingPage allows users to enable or disable their public profile.
func userPublicProfileSettingPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Public Profile"
	if _, err := cd.Queries().GetPublicProfileRoleForUser(r.Context(), cd.UserID); err != nil {
		http.NotFound(w, r)
		return
	}
	user, _ := cd.CurrentUser()
	data := struct {
		Enabled bool
	}{
		Enabled: user.PublicProfileEnabledAt.Valid,
	}
	UserPublicProfileSettingsPage.Handle(w, r, data)
}

const UserPublicProfileSettingsPage handlers.Page = "user/publicProfileSettings.gohtml"

type PublicProfileSaveTask struct{ tasks.TaskString }

var publicProfileSaveTask = &PublicProfileSaveTask{TaskString: TaskSavePublicProfile}
var _ tasks.Task = (*PublicProfileSaveTask)(nil)

func (PublicProfileSaveTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if _, err := queries.GetPublicProfileRoleForUser(r.Context(), uid); err != nil {
		return common.UserError{ErrorMessage: "forbidden"}
	}
	enable := r.PostFormValue("enable") != ""
	var ts sql.NullTime
	if enable {
		ts = sql.NullTime{Time: time.Now(), Valid: true}
	}
	if err := queries.UpdatePublicProfileEnabledAtForUser(r.Context(), db.UpdatePublicProfileEnabledAtForUserParams{EnabledAt: ts, UserID: uid, GranteeID: sql.NullInt32{Int32: uid, Valid: uid != 0}}); err != nil {
		return fmt.Errorf("update public profile fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/profile"}
}
