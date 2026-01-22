package auth

import (
	"net/http"
	"time"

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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	code := r.FormValue("code")
	pw := r.FormValue("password")
	sig := r.FormValue("sig")
	ts := r.FormValue("ts")

	var expiry *time.Time
	if sig != "" {
		t, err := cd.VerifyPasswordResetLink(code, sig, ts)
		if err != nil {
			return handlers.ErrRedirectOnSamePageHandler(err)
		}
		expiry = &t
	}

	if err := cd.VerifyPasswordReset(code, pw, expiry); err != nil {
		return handlers.ErrRedirectOnSamePageHandler(err)
	}
	return handlers.RefreshDirectHandler{TargetURL: "/login"}
}
