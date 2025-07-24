package user

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/middleware"

	"github.com/arran4/goa4web/internal/db"
)

func userEmailPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
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
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	user, _ := cd.CurrentUser()
	pref, _ := cd.Preference()

	emails, _ := queries.GetUserEmailsByUserID(r.Context(), cd.UserID)
	var verified, unverified []*db.UserEmail
	for _, e := range emails {
		if e.VerifiedAt.Valid {
			verified = append(verified, e)
		} else {
			unverified = append(unverified, e)
		}
	}
	data := Data{
		CoreData:   r.Context().Value(consts.KeyCoreData).(*common.CoreData),
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

	handlers.TemplateHandler(w, r, "emailPage.gohtml", data)
}

func userEmailVerifyCodePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, err := cd.GetSession(r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		middleware.RedirectToLogin(w, r, session)
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
			handlers.TemplateHandler(w, r, "user/emailVerifiedPage.gohtml", struct{ *common.CoreData }{r.Context().Value(consts.KeyCoreData).(*common.CoreData)})
			return
		}
		if err := queries.UpdateUserEmailVerification(r.Context(), db.UpdateUserEmailVerificationParams{VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: ue.ID}); err != nil {
			log.Printf("update user email verification: %v", err)
		}
		if err := queries.DeleteUserEmailsByEmailExceptID(r.Context(), db.DeleteUserEmailsByEmailExceptIDParams{Email: ue.Email, ID: ue.ID}); err != nil {
			log.Printf("delete user emails: %v", err)
		}
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}

	if ue.VerifiedAt.Valid {
		handlers.TemplateHandler(w, r, "user/emailVerifiedPage.gohtml", struct{ *common.CoreData }{r.Context().Value(consts.KeyCoreData).(*common.CoreData)})
		return
	}

	data := struct {
		*common.CoreData
		Code  string
		Email string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Code:     code,
		Email:    ue.Email,
	}
	handlers.TemplateHandler(w, r, "user/emailVerifyConfirmPage.gohtml", data)
}
