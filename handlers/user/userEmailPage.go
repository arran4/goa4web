package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	handlers "github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

type SaveEmailTask struct{ tasks.TaskString }
type AddEmailTask struct{ tasks.TaskString }
type ResendEmailTask struct{ tasks.TaskString }
type DeleteEmailTask struct{ tasks.TaskString }
type TestMailTask struct{ tasks.TaskString }

var _ tasks.Task = (*TestMailTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*TestMailTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*AddEmailTask)(nil)

func (ResendEmailTask) Action(w http.ResponseWriter, r *http.Request) { addEmailTask.Resend(w, r) }

var (
	saveEmailTask   = &SaveEmailTask{TaskString: tasks.TaskString(TaskSaveAll)}
	addEmailTask    = &AddEmailTask{TaskString: tasks.TaskString(TaskAdd)}
	resendEmailTask = &ResendEmailTask{TaskString: TaskResend}
	deleteEmailTask = &DeleteEmailTask{TaskString: tasks.TaskString(TaskDelete)}
	testMailTask    = &TestMailTask{TaskString: tasks.TaskString(TaskTestMail)}
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

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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
func (SaveEmailTask) Action(w http.ResponseWriter, r *http.Request) {
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

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

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

func (TestMailTask) Action(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	user, _ := cd.CurrentUser()
	if user == nil {
		http.Error(w, "email unknown", http.StatusBadRequest)
		return
	}
	if cd.EmailProvider() == nil {
		q := url.QueryEscape(ErrMailNotConfigured.Error())
		r.URL.RawQuery = "error=" + q
		handlers.TaskErrorAcknowledgementPage(w, r)
		return
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func (TestMailTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("testEmail")
}

func (TestMailTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("testEmail")
	return &s
}

func (AddEmailTask) Action(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if ue, err := queries.GetUserEmailByEmail(r.Context(), emailAddr); err == nil && ue.VerifiedAt.Valid {
		http.Redirect(w, r, "/usr/email?error=email+exists", http.StatusSeeOther)
		return
	}
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	code := hex.EncodeToString(buf[:])
	expire := time.Now().Add(24 * time.Hour)
	if err := queries.InsertUserEmail(r.Context(), db.InsertUserEmailParams{UserID: uid, Email: emailAddr, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{String: code, Valid: true}, VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true}, NotificationPriority: 0}); err != nil {
		// TODO better error for NOT Duplicate entry  for key 'user_emails_email_idx'
		log.Printf("insert user email: %v", err)
		http.Redirect(w, r, "/usr/email?error=email+exists", http.StatusSeeOther)
		return
	}
	path := "/usr/email/verify?code=" + code
	page := "http://" + r.Host + path
	if config.AppRuntimeConfig.HTTPHostname != "" {
		page = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + path
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	evt := cd.Event()
	if evt.Data == nil {
		evt.Data = map[string]any{}
	}
	evt.Data["URL"] = page
	if user, err := cd.CurrentUser(); err == nil && user != nil {
		evt.Data["Username"] = user.Username.String
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func (AddEmailTask) Resend(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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
	path := "/usr/email/verify?code=" + code
	page := "http://" + r.Host + path
	if config.AppRuntimeConfig.HTTPHostname != "" {
		page = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/") + path
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	evt := cd.Event()
	if evt.Data == nil {
		evt.Data = map[string]any{}
	}
	evt.Data["URL"] = page
	if user, err := cd.CurrentUser(); err == nil && user != nil {
		evt.Data["Username"] = user.Username.String
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func (DeleteEmailTask) Action(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	ue, err := queries.GetUserEmailByID(r.Context(), int32(id))
	if err == nil && ue.UserID == uid {
		_ = queries.DeleteUserEmail(r.Context(), int32(id))
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func (AddEmailTask) Notify(w http.ResponseWriter, r *http.Request) {
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
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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

func (AddEmailTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("verifyEmail")
}

func (AddEmailTask) SelfInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("verifyEmail")
	return &v
}

func userEmailVerifyCodePage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.NotFound(w, r)
		return
	}
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	ue, err := queries.GetUserEmailByCode(r.Context(), sql.NullString{String: code, Valid: true})
	if err != nil || (ue.VerificationExpiresAt.Valid && ue.VerificationExpiresAt.Time.Before(time.Now())) {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	if ue.UserID != uid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	_ = queries.UpdateUserEmailVerification(r.Context(), db.UpdateUserEmailVerificationParams{VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: ue.ID})
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}
