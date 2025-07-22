package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

type ForgotPasswordTask struct {
	tasks.TaskString
}

// EmailAssociationRequestTask allows a user to request an email association.
type EmailAssociationRequestTask struct{ tasks.TaskString }

var _ tasks.Task = (*ForgotPasswordTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*ForgotPasswordTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*ForgotPasswordTask)(nil)
var _ tasks.Task = (*EmailAssociationRequestTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*EmailAssociationRequestTask)(nil)
var _ notif.SelfEmailBroadcaster = (*ForgotPasswordTask)(nil)

// ForgotPasswordTask handles password reset requests.
var forgotPasswordTask = &ForgotPasswordTask{
	TaskString: TaskUserResetPassword,
}

var emailAssociationRequestTask = &EmailAssociationRequestTask{TaskString: TaskEmailAssociationRequest}

func (EmailAssociationRequestTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationEmailAssociationRequestEmail")
}

func (EmailAssociationRequestTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationEmailAssociationRequestEmail")
	return &v
}

func (f ForgotPasswordTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationUserRequestPasswordResetEmail")
}

func (f ForgotPasswordTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationUserRequestPasswordResetEmail")
	return &v
}

func (f ForgotPasswordTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("passwordResetEmail")
}

func (f ForgotPasswordTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("password_reset")
	return &s
}

func (ForgotPasswordTask) SelfEmailBroadcast() bool { return true }

func (ForgotPasswordTask) Page(w http.ResponseWriter, r *http.Request) {
	handlers.TemplateHandler(w, r, "forgotPasswordPage.gohtml", r.Context().Value(consts.KeyCoreData))
}

func (ForgotPasswordTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	pw := r.PostFormValue("password")
	if username == "" || pw == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	row, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if row.Email == "" {
		type Data struct {
			*common.CoreData
			Username    string
			RequestTask string
		}
		data := Data{
			CoreData:    r.Context().Value(consts.KeyCoreData).(*common.CoreData),
			Username:    row.Username.String,
			RequestTask: string(TaskEmailAssociationRequest),
		}
		handlers.TemplateHandler(w, r, "forgotPasswordNoEmailPage.gohtml", data)
		return
	}

	if reset, err := queries.GetPasswordResetByUser(r.Context(), db.GetPasswordResetByUserParams{
		UserID:    row.Idusers,
		CreatedAt: time.Now().Add(-time.Duration(config.AppRuntimeConfig.PasswordResetExpiryHours) * time.Hour),
	}); err == nil {
		if time.Since(reset.CreatedAt) < 24*time.Hour {
			http.Error(w, "reset recently requested", http.StatusTooManyRequests)
			return
		}
		_ = queries.DeletePasswordReset(r.Context(), reset.ID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get reset: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	hash, alg, err := HashPassword(pw)
	if err != nil {
		http.Error(w, "hash error", http.StatusInternalServerError)
		return
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		http.Error(w, "rand", http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(buf[:])
	if err := queries.CreatePasswordReset(r.Context(), db.CreatePasswordResetParams{UserID: row.Idusers, Passwd: hash, PasswdAlgorithm: alg, VerificationCode: code}); err != nil {
		log.Printf("create reset: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if row.Email != "" {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["reset"] = notif.PasswordResetInfo{Username: row.Username.String, Code: code}
				evt.Data["ResetURL"] = cd.AbsoluteURL("/login?code=" + code)
				evt.Data["UserURL"] = cd.AbsoluteURL(fmt.Sprintf("/admin/user/%d", row.Idusers))
			}
		}
	}
	handlers.TemplateHandler(w, r, "forgotPasswordEmailSentPage.gohtml", r.Context().Value(consts.KeyCoreData))
}

func (EmailAssociationRequestTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	reason := r.PostFormValue("reason")
	if username == "" || email == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	row, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if row.Email != "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	res, err := queries.InsertAdminRequestQueue(r.Context(), db.InsertAdminRequestQueueParams{
		UsersIdusers:   row.Idusers,
		ChangeTable:    "user_emails",
		ChangeField:    "email",
		ChangeRowID:    row.Idusers,
		ChangeValue:    sql.NullString{String: email, Valid: true},
		ContactOptions: sql.NullString{String: email, Valid: true},
	})
	if err != nil {
		log.Printf("insert admin request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	_ = queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: reason})
	_ = queries.InsertAdminUserComment(r.Context(), db.InsertAdminUserCommentParams{UsersIdusers: row.Idusers, Comment: "email association requested"})
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Path = fmt.Sprintf("/admin/request/%d", id)
			evt.Data["Username"] = row.Username.String
			evt.Data["Email"] = email
			evt.Data["Reason"] = reason
			evt.Data["UserURL"] = cd.AbsoluteURL(fmt.Sprintf("/admin/user/%d", row.Idusers))
		}
	}
	handlers.TemplateHandler(w, r, "forgotPasswordRequestSentPage.gohtml", r.Context().Value(consts.KeyCoreData))
}
