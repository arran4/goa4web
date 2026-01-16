package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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

// Ensure template requirements are declared for this task.
var _ tasks.TemplatesRequired = (*ForgotPasswordTask)(nil)

const (
	ForgotPasswordPageTmpl          handlers.Page = "forgotPasswordPage.gohtml"
	ForgotPasswordNoEmailPageTmpl   handlers.Page = "forgotPasswordNoEmailPage.gohtml"
	ForgotPasswordEmailSentPageTmpl handlers.Page = "forgotPasswordEmailSentPage.gohtml"
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
	row, err := cd.UserCredentials(username)
	if err != nil {
		return fmt.Errorf("user not found %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	verifiedEmails, err := cd.VerifiedEmailsForUser(row.Idusers)
	if err != nil {
		return fmt.Errorf("user email list %w", err)
	}
	if _, err := queries.GetLoginRoleForUser(r.Context(), row.Idusers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.ErrRedirectOnSamePageHandler(errors.New("approval is pending"))
		}
		return fmt.Errorf("user role %w", err)
	}
	userHasNoVerifiedEmail := len(verifiedEmails) == 0
	hash, alg, err := HashPassword(pw)
	if err != nil {
		return fmt.Errorf("hash error %w", err)
	}

	var code string
	if userHasNoVerifiedEmail {
		code, err = cd.CreatePasswordResetForUser(row.Idusers, hash, alg)
	} else {
		code, err = cd.CreatePasswordReset(verifiedEmails[0], hash, alg)
	}
	if err != nil {
		if errors.Is(err, common.ErrPasswordResetRecentlyRequested) {
			return handlers.ErrRedirectOnSamePageHandler(err)
		}
		log.Printf("create reset: %v", err)
		return fmt.Errorf("create reset %w", err)
	}
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
			evt.Data["Emails"] = verifiedEmails
			evt.Data["UserHasNoVerifiedEmail"] = userHasNoVerifiedEmail
		}
	}
	cd.PageTitle = "Password Reset"
	return ForgotPasswordEmailSentPageTmpl.Handler(struct{}{})
}

func (ForgotPasswordTask) AuditRecord(data map[string]any) string {
	if u, ok := data["Username"].(string); ok {
		return "password reset requested for " + u
	}
	return "password reset requested"
}

func (EmailAssociationRequestTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationEmailAssociationRequestEmail"), true
}

func (EmailAssociationRequestTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationEmailAssociationRequestEmail")
	return &v
}

func (f ForgotPasswordTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminNotificationUserRequestPasswordResetEmail"), true
}

func (f ForgotPasswordTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationUserRequestPasswordResetEmail")
	return &v
}

func (f ForgotPasswordTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	if v, ok := evt.Data["UserHasNoVerifiedEmail"].(bool); ok && v {
		return nil, false
	}
	return notif.NewEmailTemplates("passwordResetEmail"), true
}

func (f ForgotPasswordTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("password_reset")
	return &s
}

func (ForgotPasswordTask) SelfEmailBroadcast() bool { return true }

func (ForgotPasswordTask) Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Password Reset"
	ForgotPasswordPageTmpl.Handle(w, r, struct{}{})
}

// TemplatesRequired declares templates used by ForgotPasswordTask.
func (ForgotPasswordTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{
		ForgotPasswordPageTmpl,
		ForgotPasswordNoEmailPageTmpl,
		ForgotPasswordEmailSentPageTmpl,
	}
}
