package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserPasswordResetTask resets a user's password and notifies them.
type UserPasswordResetTask struct{ tasks.TaskString }

var userPasswordResetTask = &UserPasswordResetTask{TaskString: TaskUserResetPassword}

const (
	TemplateUserResetPasswordConfirmPage handlers.Page = "admin/userResetPasswordConfirmPage.gohtml"
)

var _ tasks.Task = (*UserPasswordResetTask)(nil)
var _ tasks.AuditableTask = (*UserPasswordResetTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserPasswordResetTask)(nil)
var _ tasks.TemplatesRequired = (*UserPasswordResetTask)(nil)

func (UserPasswordResetTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["targetUserID"] = user.Idusers
		evt.Data["Username"] = user.Username.String
		evt.Data["Password"] = newPass
	}
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func (UserPasswordResetTask) TemplatesRequired() []string {
	return []string{
		handlers.TemplateRunTaskPage,
		string(TemplateUserResetPasswordConfirmPage),
	}
}

func (UserPasswordResetTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if id, ok := evt.Data["targetUserID"].(int32); ok {
		return []int32{id}, nil
	}
	if id, ok := evt.Data["targetUserID"].(int); ok {
		return []int32{int32(id)}, nil
	}
	return nil, fmt.Errorf("target user id not provided")
}

func (UserPasswordResetTask) TargetEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("adminPasswordResetEmail"), true
}

func (UserPasswordResetTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := notif.NotificationTemplateFilenameGenerator("admin_password_reset")
	return &v
}

func (UserPasswordResetTask) AuditRecord(data map[string]any) string {
	if u, ok := data["Username"].(string); ok {
		return "password reset for " + u
	}
	if id, ok := data["targetUserID"].(int32); ok {
		return fmt.Sprintf("password reset for %d", id)
	}
	return "password reset"
}

func adminUserResetPasswordConfirmPage(w http.ResponseWriter, r *http.Request) {
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
	TemplateUserResetPasswordConfirmPage.Handle(w, r, data)
}
