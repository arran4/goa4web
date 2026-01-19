package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminGenerateResetURLTask struct{ tasks.TaskString }

var adminGenerateResetURLTask = &AdminGenerateResetURLTask{TaskString: "admin_generate_reset_url"}

var _ tasks.Task = (*AdminGenerateResetURLTask)(nil)
var _ tasks.TemplatesRequired = (*AdminGenerateResetURLTask)(nil)
var _ notif.TargetUsersNotificationProvider = (*AdminGenerateResetURLTask)(nil)

const TemplateAdminUserGenerateResetPage handlers.Page = "admin/userGenerateResetPage.gohtml"

func (AdminGenerateResetURLTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return nil
	}

	code, err := cd.CreatePasswordResetTokenForUser(user.Idusers)
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	resetURL := cd.AbsoluteURL(fmt.Sprintf("/reset?code=%s", code))

	// Send internal notification if possible
	if evt := cd.Event(); evt != nil {
		evt.UserID = user.Idusers
		evt.Data = map[string]any{
			"ResetURL": resetURL,
			"User":     user,
		}
	}

	data := struct {
		User     *db.SystemGetUserByIDRow
		ResetURL string
		Back     string
	}{
		User:     user,
		ResetURL: resetURL,
		Back:     fmt.Sprintf("/admin/user/%d", user.Idusers),
	}

	return handlers.TemplateWithDataHandler(TemplateAdminUserGenerateResetPage, data)
}

func (AdminGenerateResetURLTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{TemplateAdminUserGenerateResetPage}
}

func (AdminGenerateResetURLTask) TargetUserIDs(evt eventbus.TaskEvent) ([]int32, error) {
	if user, ok := evt.Data["User"].(*db.SystemGetUserByIDRow); ok {
		return []int32{user.Idusers}, nil
	}
	return nil, fmt.Errorf("user not found in event data")
}

func (AdminGenerateResetURLTask) TargetEmailTemplate(evt eventbus.TaskEvent) (*notif.EmailTemplates, bool) {
	return notif.NewEmailTemplates("passwordResetEmail"), true
}

func (AdminGenerateResetURLTask) TargetInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	t := "adminGeneratedResetURL"
	return &t
}
