package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// ResetPasswordPage handles the GET request for password reset links.
func ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	code := r.FormValue("code")
	sig := r.FormValue("sig")
	ts := r.FormValue("ts") // or expires param if that's what we decided. SignPasswordResetLink uses WithExpiry which adds "ts" or "expires"?
	// coredata_auth.go: SignPasswordResetLink uses sign.AddQuerySig(..., sign.WithExpiry(exp))
	// internal/sign/sign.go: AddQuerySig adds "ts" for WithExpiry.

	// Verify signature first
	expires, err := cd.VerifyPasswordResetLink(code, sig, ts)
	if err != nil {
		fmt.Printf("DEBUG: Verify failed. Code=%q Sig=%q TS=%q Err=%v\n", code, sig, ts, err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid or expired link: %w", err))
		return
	}

	cd.PageTitle = "Reset Password"

	data := struct {
		Code    string
		Sig     string
		TS      string
		Expires time.Time
	}{
		Code:    code,
		Sig:     sig,
		TS:      ts,
		Expires: expires,
	}

	TemplateResetPasswordPage.Handle(w, r, data)
}

const TemplateResetPasswordPage handlers.Page = "resetPasswordPage.gohtml"
