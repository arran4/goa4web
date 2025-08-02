package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// EmailAssociationRequestTask allows a user to request an email association.
type EmailAssociationRequestTask struct{ tasks.TaskString }

var (
	_ tasks.Task                       = (*EmailAssociationRequestTask)(nil)
	_ tasks.AuditableTask              = (*EmailAssociationRequestTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*EmailAssociationRequestTask)(nil)
)

var emailAssociationRequestTask = &EmailAssociationRequestTask{TaskString: TaskEmailAssociationRequest}

func (EmailAssociationRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"username", "email", "reason"}, []string{"username", "email"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	reason := r.PostFormValue("reason")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	row, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		return fmt.Errorf("user not found %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if row.Email != "" {
		return handlers.RefreshDirectHandler{TargetURL: "/login"}
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
		return fmt.Errorf("insert admin request %w", err)
	}
	id, _ := res.LastInsertId()
	_ = queries.InsertAdminRequestComment(r.Context(), db.InsertAdminRequestCommentParams{RequestID: int32(id), Comment: reason})
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.TemplateHandler(w, r, "forgotPasswordRequestSentPage.gohtml", r.Context().Value(consts.KeyCoreData))
	})
}

func (EmailAssociationRequestTask) AuditRecord(data map[string]any) string {
	u, _ := data["Username"].(string)
	e, _ := data["Email"].(string)
	if u != "" && e != "" {
		return fmt.Sprintf("email association request for %s -> %s", u, e)
	}
	return "email association request"
}
