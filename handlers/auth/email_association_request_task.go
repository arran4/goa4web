package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// EmailAssociationRequestTask allows a user to request an email association.
type EmailAssociationRequestTask struct{ tasks.TaskString }

var (
	_ tasks.Task                       = (*EmailAssociationRequestTask)(nil)
	_ tasks.AuditableTask              = (*EmailAssociationRequestTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*EmailAssociationRequestTask)(nil)
	_ tasks.EmailTemplatesRequired     = (*EmailAssociationRequestTask)(nil)
)

var emailAssociationRequestTask = &EmailAssociationRequestTask{TaskString: TaskEmailAssociationRequest}

func (EmailAssociationRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := handlers.ValidateForm(r, []string{"username", "email", "reason"}, []string{"username", "email"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	reason := r.PostFormValue("reason")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	params := common.AssociateEmailParams{Username: username, Email: email, Reason: reason}
	row, id, err := cd.AssociateEmail(params)
	if err != nil {
		if errors.Is(err, common.ErrEmailAlreadyAssociated) {
			return handlers.RefreshDirectHandler{TargetURL: "/login"}
		}
		return fmt.Errorf("associate email %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ForgotPasswordRequestSentPageTmpl.Handle(w, r, struct{}{})
	})
}

const ForgotPasswordRequestSentPageTmpl handlers.Page = "forgotPasswordRequestSentPage.gohtml"

func (EmailAssociationRequestTask) AuditRecord(data map[string]any) string {
	u, _ := data["Username"].(string)
	e, _ := data["Email"].(string)
	if u != "" && e != "" {
		return fmt.Sprintf("email association request for %s -> %s", u, e)
	}
	return "email association request"
}
