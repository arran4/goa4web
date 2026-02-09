package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserResetPasswordTask handles the public password reset via magic link.
type UserResetPasswordTask struct{ tasks.TaskString }

var userResetPasswordTask = &UserResetPasswordTask{TaskString: "Password Reset"}

const TemplateUserResetPasswordPage tasks.Template = "userResetPasswordPage.gohtml"

var _ tasks.Task = (*UserResetPasswordTask)(nil)
var _ tasks.TemplatesRequired = (*UserResetPasswordTask)(nil)

func (UserResetPasswordTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	idStr := vars["user"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid user id"))
	}

	sig := r.FormValue("sig")
	ts := r.FormValue("ts")
	code := r.FormValue("code")

	// Reconstruct signed data: link:/user/{id}/reset?code={code}
	path := fmt.Sprintf("/user/%d/reset?code=%s", id, code)
	data := "link:" + path

	// Verify signature
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid timestamp: %w", err))
	}
	if err := sign.Verify(data, sig, cd.LinkSignKey, sign.WithExpiry(time.Unix(tsInt, 0))); err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid or expired link: %w", err))
	}

	// Verify token in DB
	queries := cd.Queries()
	// Use config expiry? Or signature expiry? Signature expiry covers time.
	// But we also need to check DB validity.
	reset, err := queries.GetPasswordResetByCode(r.Context(), db.GetPasswordResetByCodeParams{
		VerificationCode: code,
		CreatedAt:        time.Now().Add(-24 * time.Hour), // Ensure recent enough
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid or expired token"))
		}
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("db error: %w", err))
	}
	if reset.UserID != int32(id) {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("token user mismatch"))
	}
	if reset.PasswdAlgorithm != "magic" {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid token type"))
	}

	password := r.PostFormValue("password")
	confirm := r.PostFormValue("confirm_password")

	if password == "" {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("password required"))
	}
	if password != confirm {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("passwords do not match"))
	}

	hash, alg, err := auth.HashPassword(password)
	if err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("hash error: %w", err))
	}

	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{
		UsersIdusers:    int32(id),
		Passwd:          hash,
		PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
	}); err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("update password: %w", err))
	}

	// Invalidate token
	if err := queries.SystemMarkPasswordResetVerified(r.Context(), reset.ID); err != nil {
		// Log error but proceed
	}

	return handlers.RefreshDirectHandler{TargetURL: "/login?notice=Password+updated.+Please+login."}
}

func (UserResetPasswordTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{TemplateUserResetPasswordPage}
}

func UserResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Set New Password"

	vars := mux.Vars(r)
	idStr := vars["user"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid user id"))
		return
	}

	sig := r.URL.Query().Get("sig")
	ts := r.URL.Query().Get("ts")
	code := r.URL.Query().Get("code")

	path := fmt.Sprintf("/user/%d/reset?code=%s", id, code)
	data := "link:" + path

	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid timestamp"))
		return
	}

	if err := sign.Verify(data, sig, cd.LinkSignKey, sign.WithExpiry(time.Unix(tsInt, 0))); err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid or expired link"))
		return
	}

	// Check DB validity on GET too? Good practice.
	queries := cd.Queries()
	reset, err := queries.GetPasswordResetByCode(r.Context(), db.GetPasswordResetByCodeParams{
		VerificationCode: code,
		CreatedAt:        time.Now().Add(-24 * time.Hour),
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid or expired token"))
		return
	}
	if reset.UserID != int32(id) || reset.PasswdAlgorithm != "magic" {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid token"))
		return
	}

	// Pass code, sig, ts to template?
	// If form action is empty, browser submits to current URL (with query params).
	// So we don't strictly need to pass them to hidden fields if we rely on that.
	// But it is safer to be explicit if action is set.
	// We'll rely on empty action submitting query params.

	TemplateUserResetPasswordPage.Handle(w, r, struct {
		Sig  string
		Ts   string
		Code string
	}{
		Sig:  sig,
		Ts:   ts,
		Code: code,
	})
}
