package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

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

var _ tasks.Task = (*UserPasswordResetTask)(nil)
var _ tasks.AuditableTask = (*UserPasswordResetTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*UserPasswordResetTask)(nil)

func (UserPasswordResetTask) Action(w http.ResponseWriter, r *http.Request) any {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: cd,
		Back:     "/admin/user/" + idStr,
	}
	userRow, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		data.Errors = append(data.Errors, "user not found")
		return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rand: %w", err).Error())
		return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
	}
	newPass := hex.EncodeToString(buf[:])
	hash, alg, err := auth.HashPassword(newPass)
	if err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("hashPassword: %w", err).Error())
		return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
	}
	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: int32(id), Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: true}}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("reset password: %w", err).Error())
		return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["targetUserID"] = int32(id)
		evt.Data["Username"] = userRow.Username.String
		evt.Data["Password"] = newPass
	}
	return handlers.TemplateWithDataHandler("runTaskPage.gohtml", data)
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

func (UserPasswordResetTask) TargetEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminPasswordResetEmail")
}

func (UserPasswordResetTask) TargetInternalNotificationTemplate() *string {
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
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	userRow, err := queries.GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	data := struct {
		*common.CoreData
		User *db.User
		Back string
	}{
		CoreData: cd,
		User:     &db.User{Idusers: userRow.Idusers, Username: userRow.Username},
		Back:     "/admin/user/" + idStr,
	}
	handlers.TemplateHandler(w, r, "userResetPasswordConfirmPage.gohtml", data)
}
