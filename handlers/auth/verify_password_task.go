package auth

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// VerifyPasswordTask verifies reset codes during login.
type VerifyPasswordTask struct {
	tasks.TaskString
}

// verifyPasswordTask handles password verification requests.
var verifyPasswordTask = &VerifyPasswordTask{TaskString: TaskPasswordVerify}

// ensure VerifyPasswordTask conforms to tasks.Task
var _ tasks.Task = (*VerifyPasswordTask)(nil)

// Action verifies a password reset code and logs the user in.
func (VerifyPasswordTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return handlers.RefreshDirectHandler{TargetURL: "/login"}
	}
	code := r.FormValue("code")
	pw := r.FormValue("password")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.VerifyPasswordReset(code, pw); err != nil {
		return handlers.ErrRedirectOnSamePageHandler(err)
	}
	return handlers.RefreshDirectHandler{TargetURL: "/login"}
}
