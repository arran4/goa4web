package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/utils/emailutil"

	"github.com/arran4/goa4web/config"
)

// ErrMailNotConfigured is returned when test mail has no provider configured.
var ErrMailNotConfigured = errors.New("mail isn't configured")

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

	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
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
		CoreData:   r.Context().Value(common.KeyCoreData).(*common.CoreData),
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

	if err := templates.RenderTemplate(w, "emailPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("user email page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func userEmailSaveActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)

	updates := r.PostFormValue("emailupdates") != ""
	auto := r.PostFormValue("autosubscribe") != ""

	_, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("preference load: %v", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = queries.InsertEmailPreference(r.Context(), db.InsertEmailPreferenceParams{
				Emailforumupdates:    sql.NullBool{Bool: updates, Valid: true},
				AutoSubscribeReplies: auto,
				UsersIdusers:         uid,
			})
		}
	} else {
		err = queries.UpdateEmailForumUpdatesByUserID(r.Context(), db.UpdateEmailForumUpdatesByUserIDParams{
			Emailforumupdates: sql.NullBool{Bool: updates, Valid: true},
			UsersIdusers:      uid,
		})
		if err == nil {
			err = queries.UpdateAutoSubscribeRepliesByUserID(r.Context(), db.UpdateAutoSubscribeRepliesByUserIDParams{
				AutoSubscribeReplies: auto,
				UsersIdusers:         uid,
			})
		}
	}
	if err != nil {
		log.Printf("save email pref: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailTestActionPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	user, _ := cd.CurrentUser()
	if user == nil {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	queries, _ := r.Context().Value(common.KeyQueries).(*db.Queries)
	var emails []*db.UserEmail
	var err error
	if queries != nil {
		emails, err = queries.GetUserEmailsByUserID(r.Context(), user.Idusers)
	}
	if err != nil || len(emails) == 0 {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	addr := emails[0].Email
	base := "http://" + r.Host
	if config.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/")
	}
	pageURL := base + r.URL.Path
	provider := email.ProviderFromConfig(config.AppRuntimeConfig)
	if provider == nil {
		q := url.QueryEscape(ErrMailNotConfigured.Error())
		// Display the error without redirecting so the POST isn't repeated.
		r.URL.RawQuery = "error=" + q
		common.TaskErrorAcknowledgementPage(w, r)
		return
	}
	if err := emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, user.Idusers, addr, pageURL, "update", nil); err != nil {
		log.Printf("notify Error: %s", err)
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailAddActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	emailAddr := r.FormValue("new_email")
	if emailAddr == "" {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if ue, err := queries.GetUserEmailByEmail(r.Context(), emailAddr); err == nil && ue.VerifiedAt.Valid {
		http.Redirect(w, r, "/usr/email?error=email+exists", http.StatusSeeOther)
		return
	}
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	code := hex.EncodeToString(buf[:])
	expire := time.Now().Add(24 * time.Hour)
	_ = queries.InsertUserEmail(r.Context(), db.InsertUserEmailParams{UserID: uid, Email: emailAddr, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{String: code, Valid: true}, VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true}, NotificationPriority: 0})
	page := "http://" + r.Host + "/usr/email/verify?code=" + code
	if config.AppRuntimeConfig.HTTPHostname != "" {
		page = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + "/usr/email/verify?code=" + code
	}
	_ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, uid, emailAddr, page, common.TaskUserEmailVerification, nil)
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailResendActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	ue, err := queries.GetUserEmailByID(r.Context(), int32(id))
	if err != nil || ue.UserID != uid {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	code := hex.EncodeToString(buf[:])
	expire := time.Now().Add(24 * time.Hour)
	_ = queries.SetVerificationCode(r.Context(), db.SetVerificationCodeParams{LastVerificationCode: sql.NullString{String: code, Valid: true}, VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true}, ID: int32(id)})
	page := "http://" + r.Host + "/usr/email/verify?code=" + code
	if config.AppRuntimeConfig.HTTPHostname != "" {
		page = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + "/usr/email/verify?code=" + code
	}
	_ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, uid, ue.Email, page, common.TaskUserEmailVerification, nil)
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	ue, err := queries.GetUserEmailByID(r.Context(), int32(id))
	if err == nil && ue.UserID == uid {
		_ = queries.DeleteUserEmail(r.Context(), int32(id))
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailNotifyActionPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	val, _ := queries.GetMaxNotificationPriority(r.Context(), uid)
	var maxPr int32
	switch v := val.(type) {
	case int64:
		maxPr = int32(v)
	case int32:
		maxPr = v
	}
	_ = queries.SetNotificationPriority(r.Context(), db.SetNotificationPriorityParams{NotificationPriority: maxPr + 1, ID: int32(id)})
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func userEmailVerifyCodePage(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.NotFound(w, r)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	ue, err := queries.GetUserEmailByCode(r.Context(), sql.NullString{String: code, Valid: true})
	if err != nil || (ue.VerificationExpiresAt.Valid && ue.VerificationExpiresAt.Time.Before(time.Now())) {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	_ = queries.UpdateUserEmailVerification(r.Context(), db.UpdateUserEmailVerificationParams{VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: ue.ID})
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}
