package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// ForgotPasswordTask handles password reset requests.
type ForgotPasswordTask struct {
	tasks.TaskString
}

var (
	_ tasks.Task                             = (*ForgotPasswordTask)(nil)
	_ tasks.AuditableTask                    = (*ForgotPasswordTask)(nil)
	_ notif.SelfNotificationTemplateProvider = (*ForgotPasswordTask)(nil)
	_ notif.AdminEmailTemplateProvider       = (*ForgotPasswordTask)(nil)
	_ notif.SelfEmailBroadcaster             = (*ForgotPasswordTask)(nil)
)

var forgotPasswordTask = &ForgotPasswordTask{TaskString: TaskUserResetPassword}

func (ForgotPasswordTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"username", "password"}, []string{"username", "password"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	username := r.PostFormValue("username")
	pw := r.PostFormValue("password")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	row, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		return fmt.Errorf("user not found %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if _, err := queries.GetLoginRoleForUser(r.Context(), row.Idusers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.ErrRedirectOnSamePageHandler(errors.New("approval is pending"))
		}
		return fmt.Errorf("user role %w", err)
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
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			cd.PageTitle = "Password Reset"
			handlers.TemplateHandler(w, r, "forgotPasswordNoEmailPage.gohtml", data)
		})
	}

	if reset, err := queries.GetPasswordResetByUser(r.Context(), db.GetPasswordResetByUserParams{
		UserID:    row.Idusers,
		CreatedAt: time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour),
	}); err == nil {
		if time.Since(reset.CreatedAt) < 24*time.Hour {
			return handlers.ErrRedirectOnSamePageHandler(errors.New("reset recently requested"))
		}
		_ = queries.SystemDeletePasswordReset(r.Context(), reset.ID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get reset: %v", err)
		return fmt.Errorf("get reset %w", err)
	}
	hash, alg, err := HashPassword(pw)
	if err != nil {
		return fmt.Errorf("hash error %w", err)
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Errorf("rand %w", err)
	}
	code := hex.EncodeToString(buf[:])
	if err := queries.CreatePasswordResetForUser(r.Context(), db.CreatePasswordResetForUserParams{UserID: row.Idusers, Passwd: hash, PasswdAlgorithm: alg, VerificationCode: code}); err != nil {
		log.Printf("create reset: %v", err)
		return fmt.Errorf("create reset %w", err)
	}
	if row.Email != "" {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.UserID = row.Idusers
				// Expose fields directly for email templates
				evt.Data["Username"] = row.Username.String
				evt.Data["Code"] = code
				evt.Data["ResetURL"] = cd.AbsoluteURL("/login?code=" + code)
				evt.Data["UserURL"] = cd.AbsoluteURL(fmt.Sprintf("/admin/user/%d", row.Idusers))
			}
		}
	}
	cd.PageTitle = "Password Reset"
	return handlers.TemplateWithDataHandler("forgotPasswordEmailSentPage.gohtml", cd)
}

func (ForgotPasswordTask) AuditRecord(data map[string]any) string {
	if u, ok := data["Username"].(string); ok {
		return "password reset requested for " + u
	}
	return "password reset requested"
}

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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Password Reset"
	handlers.TemplateHandler(w, r, "forgotPasswordPage.gohtml", struct{}{})
}
