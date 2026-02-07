package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserForcePasswordChangeTask resets a user's password and notifies them.
type UserForcePasswordChangeTask struct{ tasks.TaskString }

var userForcePasswordChangeTask = &UserForcePasswordChangeTask{TaskString: TaskUserForcePasswordChange}

// UserSendResetEmailTask sends a password reset email to the user.
type UserSendResetEmailTask struct{ tasks.TaskString }

var userSendResetEmailTask = &UserSendResetEmailTask{TaskString: TaskUserSendResetEmail}

// UserGenerateResetLinkTask generates a password reset link for the user.
type UserGenerateResetLinkTask struct{ tasks.TaskString }

var userGenerateResetLinkTask = &UserGenerateResetLinkTask{TaskString: TaskUserGenerateResetLink}

const (
	TemplateUserResetPasswordConfirmPage tasks.Template          = "admin/userResetPasswordConfirmPage.gohtml"
	EmailTemplateUserMagicReset          notif.EmailTemplateName = "userMagicResetEmail"
)

var _ tasks.Task = (*UserForcePasswordChangeTask)(nil)
var _ tasks.AuditableTask = (*UserForcePasswordChangeTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserForcePasswordChangeTask)(nil)
var _ tasks.TemplatesRequired = (*UserForcePasswordChangeTask)(nil)
var _ tasks.EmailTemplatesRequired = (*UserForcePasswordChangeTask)(nil)

var _ tasks.Task = (*UserSendResetEmailTask)(nil)
var _ tasks.AuditableTask = (*UserSendResetEmailTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserSendResetEmailTask)(nil)
var _ tasks.TemplatesRequired = (*UserSendResetEmailTask)(nil)
var _ tasks.EmailTemplatesRequired = (*UserSendResetEmailTask)(nil)

var _ tasks.Task = (*UserGenerateResetLinkTask)(nil)
var _ tasks.AuditableTask = (*UserGenerateResetLinkTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserGenerateResetLinkTask)(nil)
var _ tasks.TemplatesRequired = (*UserGenerateResetLinkTask)(nil)
var _ tasks.EmailTemplatesRequired = (*UserGenerateResetLinkTask)(nil)

func (UserForcePasswordChangeTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	back := "/admin/user"
	if user != nil {
		back = fmt.Sprintf("/admin/user/%d", user.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	if user == nil {
		data.Errors = append(data.Errors, "user not found")
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	queries := cd.Queries()
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rand: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	newPass := hex.EncodeToString(buf[:])
	hash, alg, err := auth.HashPassword(newPass)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("hashPassword: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: user.Idusers, Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: true}}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("reset password: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	if _, err := queries.SystemDeletePasswordResetsByUser(r.Context(), user.Idusers); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("clear password resets: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["targetUserID"] = user.Idusers
		evt.Data["Username"] = user.Username.String
		evt.Data["Password"] = newPass
	}
	data.Messages = append(data.Messages, fmt.Sprintf("Password reset to: %s", newPass))
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func (UserForcePasswordChangeTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{
		tasks.Template(handlers.TemplateRunTaskPage),
		TemplateUserResetPasswordConfirmPage,
	}, EmailTemplateAdminUserRequestPasswordReset.RequiredTemplates()...)
}

func (UserForcePasswordChangeTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserForcePasswordChangeTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminPasswordReset.EmailTemplates(), true
}

func (UserForcePasswordChangeTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminPasswordReset.NotificationTemplate()
	return &v
}

func (UserForcePasswordChangeTask) AuditRecord(data map[string]any) string {
	if u, ok := data["Username"].(string); ok {
		return "password reset for " + u
	}
	if id, ok := data["targetUserID"].(int32); ok {
		return fmt.Sprintf("password reset for %d", id)
	}
	return "password reset"
}

func (UserSendResetEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	back := "/admin/user"
	if user != nil {
		back = fmt.Sprintf("/admin/user/%d", user.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	if user == nil {
		data.Errors = append(data.Errors, "user not found")
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}

	code, err := cd.CreatePasswordResetForUser(user.Idusers, "magic-link", "magic")
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("create reset token: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}

	targetURL := fmt.Sprintf("/user/%d/reset", user.Idusers)
	// We sign the path (without code) so the signature is stable?
	// No, we should sign the full target including code to prevent tampering with code?
	// But AddQuerySig appends query params.
	// Let's include code in the base URL before signing.

	targetURLWithCode := fmt.Sprintf("%s?code=%s", targetURL, code)
	linkData := "link:" + targetURLWithCode
	duration := time.Hour * 24
	opts := []sign.SignOption{
		sign.WithExpiry(time.Now().Add(duration)),
	}
	sig := sign.Sign(linkData, cd.LinkSignKey, opts...)
	signedURL, err := sign.AddQuerySig(cd.AbsoluteURL(targetURLWithCode), sig, opts...)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("sign url: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["targetUserID"] = user.Idusers
		evt.Data["Username"] = user.Username.String
		evt.Data["ResetURL"] = signedURL
	}
	data.Messages = append(data.Messages, fmt.Sprintf("Reset email sent to %s", user.Username.String))
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func (UserSendResetEmailTask) RequiredTemplates() []tasks.Template {
	return append([]tasks.Template{
		tasks.Template(handlers.TemplateRunTaskPage),
		TemplateUserResetPasswordConfirmPage,
	}, EmailTemplateUserMagicReset.RequiredTemplates()...)
}

func (UserSendResetEmailTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserSendResetEmailTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateUserMagicReset.EmailTemplates(), true
}

func (UserSendResetEmailTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	// No internal notification for email send? Or maybe reuse one?
	// The user gets the email. Maybe admin wants a log?
	// We'll return nil for now or create a template if needed.
	// But AuditRecord covers admin log.
	return nil
}

func (UserSendResetEmailTask) AuditRecord(data map[string]any) string {
	if u, ok := data["Username"].(string); ok {
		return "password reset email sent for " + u
	}
	if id, ok := data["targetUserID"].(int32); ok {
		return fmt.Sprintf("password reset email sent for %d", id)
	}
	return "password reset email sent"
}

func (UserGenerateResetLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	back := "/admin/user"
	if user != nil {
		back = fmt.Sprintf("/admin/user/%d", user.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	if user == nil {
		data.Errors = append(data.Errors, "user not found")
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}

	code, err := cd.CreatePasswordResetForUser(user.Idusers, "magic-link", "magic")
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("create reset token: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}

	targetURL := fmt.Sprintf("/user/%d/reset", user.Idusers)
	targetURLWithCode := fmt.Sprintf("%s?code=%s", targetURL, code)
	linkData := "link:" + targetURLWithCode
	duration := time.Hour * 24
	opts := []sign.SignOption{
		sign.WithExpiry(time.Now().Add(duration)),
	}
	sig := sign.Sign(linkData, cd.LinkSignKey, opts...)
	signedURL, err := sign.AddQuerySig(cd.AbsoluteURL(targetURLWithCode), sig, opts...)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("sign url: %w", err).Error())
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["targetUserID"] = user.Idusers
		evt.Data["Username"] = user.Username.String
		evt.Data["ResetURL"] = signedURL
	}
	data.Messages = append(data.Messages, fmt.Sprintf("Reset link generated: %s", signedURL))
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func (UserGenerateResetLinkTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{
		tasks.Template(handlers.TemplateRunTaskPage),
		TemplateUserResetPasswordConfirmPage,
	}
}

func (UserGenerateResetLinkTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserGenerateResetLinkTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	// No email sent
	return nil, false
}

func (UserGenerateResetLinkTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	return nil
}

func (UserGenerateResetLinkTask) AuditRecord(data map[string]any) string {
	if u, ok := data["Username"].(string); ok {
		return "password reset link generated for " + u
	}
	if id, ok := data["targetUserID"].(int32); ok {
		return fmt.Sprintf("password reset link generated for %d", id)
	}
	return "password reset link generated"
}

type AdminUserResetPasswordConfirmPage struct{}

func (p *AdminUserResetPasswordConfirmPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Reset Password"
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	data := struct {
		User *db.User
		Back string
	}{
		User: &db.User{Idusers: user.Idusers, Username: user.Username},
		Back: fmt.Sprintf("/admin/user/%d", user.Idusers),
	}
	TemplateUserResetPasswordConfirmPage.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUserResetPasswordConfirmPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Reset Password", "", &AdminUserProfilePage{}
}

func (p *AdminUserResetPasswordConfirmPage) PageTitle() string {
	return "Reset Password"
}

var _ common.Page = (*AdminUserResetPasswordConfirmPage)(nil)
var _ http.Handler = (*AdminUserResetPasswordConfirmPage)(nil)
