package auth

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
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
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	// TODO avoid using sessions for "conversational" or "short term data storage" instead store it as query args or post if it's not sensitive
	id, _ := session.Values["PendingResetID"].(int32)
	if id == 0 {
		return handlers.RefreshDirectHandler{TargetURL: "/login"}
	}
	if err := r.ParseForm(); err != nil {
		return handlers.RefreshDirectHandler{TargetURL: "/login"}
	}
	code := r.FormValue("code")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	expiry := time.Now().Add(-time.Duration(config.AppRuntimeConfig.PasswordResetExpiryHours) * time.Hour)
	reset, err := queries.GetPasswordResetByCode(r.Context(), db.GetPasswordResetByCodeParams{VerificationCode: code, CreatedAt: expiry})
	if err != nil || reset.ID != id {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("invalid code"))
	}
	if err := queries.MarkPasswordResetVerified(r.Context(), reset.ID); err != nil {
		log.Printf("mark reset verified: %v", err)
	}
	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: reset.UserID, Passwd: reset.Passwd, PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm, Valid: true}}); err != nil {
		log.Printf("insert password: %v", err)
	}
	delete(session.Values, "PendingResetID")
	if err := session.Save(r, w); err != nil {
		log.Printf("save session: %v", err)
	}
	return handlers.RefreshDirectHandler{TargetURL: "/login"}
}
