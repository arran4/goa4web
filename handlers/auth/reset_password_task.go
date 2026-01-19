package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type ResetPasswordTask struct {
	tasks.TaskString
}

var resetPasswordTask = &ResetPasswordTask{TaskString: "reset_password"}

var _ tasks.Task = (*ResetPasswordTask)(nil)
var _ tasks.TemplatesRequired = (*ResetPasswordTask)(nil)

const TemplateResetPasswordPage handlers.Page = "resetPasswordPage.gohtml"

func (ResetPasswordTask) Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	code := r.FormValue("code")
	if code == "" {
		handlers.RenderErrorPage(w, r, fmt.Errorf("missing code"))
		return
	}
	// Verify code exists
	expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
	_, err := cd.Queries().GetPasswordResetByCode(r.Context(), db.GetPasswordResetByCodeParams{
		VerificationCode: code,
		CreatedAt:        expiry,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid or expired code"))
		return
	}

	cd.PageTitle = "Reset Password"
	type Data struct {
		Code string
	}
	TemplateResetPasswordPage.Handle(w, r, Data{Code: code})
}

func (ResetPasswordTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"code", "password"}, []string{"code", "password"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	code := r.PostFormValue("code")
	password := r.PostFormValue("password")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
	reset, err := cd.Queries().GetPasswordResetByCode(r.Context(), db.GetPasswordResetByCodeParams{
		VerificationCode: code,
		CreatedAt:        expiry,
	})
	if err != nil {
		return fmt.Errorf("invalid or expired code")
	}

	hash, alg, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash error %w", err)
	}

	if err := cd.Queries().InsertPassword(r.Context(), db.InsertPasswordParams{
		UsersIdusers:    reset.UserID,
		Passwd:          hash,
		PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
	}); err != nil {
		return fmt.Errorf("update password %w", err)
	}

	if err := cd.Queries().SystemMarkPasswordResetVerified(r.Context(), reset.ID); err != nil {
		// log?
	}

	return handlers.RefreshDirectHandler{TargetURL: "/login?notice=Password+reset+successful"}
}

func (ResetPasswordTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{TemplateResetPasswordPage}
}
