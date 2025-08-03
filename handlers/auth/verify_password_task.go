package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/internal/db"

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
	if err := r.ParseForm(); err != nil {
		return handlers.RefreshDirectHandler{TargetURL: "/login"}
	}
	idStr := r.FormValue("id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id64 == 0 {
		return handlers.RefreshDirectHandler{TargetURL: "/login"}
	}
	id := int32(id64)
	code := r.FormValue("code")
	pw := r.FormValue("password")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
	reset, err := queries.GetPasswordResetByCode(r.Context(), db.GetPasswordResetByCodeParams{VerificationCode: code, CreatedAt: expiry})
	if err != nil || reset.ID != id {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("invalid code"))
	}
	if _, err := queries.GetHasLoginRoleForUser(r.Context(), reset.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.ErrRedirectOnSamePageHandler(errors.New("approval is pending"))
		}
		return fmt.Errorf("user role %w", err)
	}
	if !VerifyPassword(pw, reset.Passwd, reset.PasswdAlgorithm) {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("invalid password"))
	}
	if err := queries.MarkPasswordResetVerified(r.Context(), reset.ID); err != nil {
		log.Printf("mark reset verified: %v", err)
	}
	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: reset.UserID, Passwd: reset.Passwd, PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm, Valid: true}}); err != nil {
		log.Printf("insert password: %v", err)
	}
	return handlers.RefreshDirectHandler{TargetURL: "/login"}
}
