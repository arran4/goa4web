package user

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		UserData        *db.User
		Verified        []*db.UserEmail
		Unverified      []*db.UserEmail
		UserPreferences struct {
			EmailUpdates         bool
			AutoSubscribeReplies bool
		}
		Error string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Email Settings"
	user, _ := cd.CurrentUser()
	pref, _ := cd.Preference()

	emails, _ := cd.UserEmails(cd.UserID)
	var verified, unverified []*db.UserEmail
	for _, e := range emails {
		if e.VerifiedAt.Valid {
			verified = append(verified, e)
		} else {
			unverified = append(unverified, e)
		}
	}
	data := Data{
		UserData:   user,
		Verified:   verified,
		Unverified: unverified,
		Error:      r.URL.Query().Get("error"),
	}
	if pref != nil {
		if pref.Emailforumupdates.Valid {
			data.UserPreferences.EmailUpdates = pref.Emailforumupdates.Bool
		}
		data.UserPreferences.AutoSubscribeReplies = pref.AutoSubscribeReplies
	} else {
		data.UserPreferences.AutoSubscribeReplies = true
	}

	UserEmailPage.Handle(w, r, data)
}

const UserEmailPage tasks.Template = "user/emailPage.gohtml"

func userEmailVerifyCodePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Verify Email"
	session, err := core.GetSession(r)
	if err != nil {
		log.Printf("get session: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		_ = middleware.RedirectToLogin(w, r, session)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		code = r.URL.Query().Get("code")
	}
	if code == "" {
		http.NotFound(w, r)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	ue, err := queries.GetUserEmailByCode(r.Context(), sql.NullString{String: code, Valid: true})
	if err != nil || (ue.VerificationExpiresAt.Valid && ue.VerificationExpiresAt.Time.Before(time.Now())) || ue.UserID != uid {
		w.WriteHeader(http.StatusNotFound)
		r.URL.RawQuery = "error=" + url.QueryEscape("Invalid verification link")
		handlers.TaskErrorAcknowledgementPage(w, r)
		return
	}

	if r.Method == http.MethodPost {
		if ue.VerifiedAt.Valid {
			UserEmailVerifiedPage.Handle(w, r, struct{}{})
			return
		}
		if err := queries.SystemMarkUserEmailVerified(r.Context(), db.SystemMarkUserEmailVerifiedParams{VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: ue.ID}); err != nil {
			log.Printf("update user email verification: %v", err)
		}
		if err := cd.AddEmail(uid, ue.ID); err != nil {
			log.Printf("promote verified email: %v", err)
		}
		if err := queries.SystemDeleteUserEmailsByEmailExceptID(r.Context(), db.SystemDeleteUserEmailsByEmailExceptIDParams{Email: ue.Email, ID: ue.ID}); err != nil {
			log.Printf("delete user emails: %v", err)
		}
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}

	if ue.VerifiedAt.Valid {
		UserEmailVerifiedPage.Handle(w, r, struct{}{})
		return
	}

	data := struct {
		Code  string
		Email string
	}{
		Code:  code,
		Email: ue.Email,
	}
	UserEmailVerifyConfirmPage.Handle(w, r, data)
}

const UserEmailVerifyConfirmPage tasks.Template = "user/emailVerifyConfirmPage.gohtml"
const UserEmailVerifiedPage tasks.Template = "user/emailVerifiedPage.gohtml"
