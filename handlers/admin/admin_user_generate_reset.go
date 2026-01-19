package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminGenerateResetURLTask struct{ tasks.TaskString }

var adminGenerateResetURLTask = &AdminGenerateResetURLTask{TaskString: "admin_generate_reset_url"}

var _ tasks.Task = (*AdminGenerateResetURLTask)(nil)
var _ tasks.TemplatesRequired = (*AdminGenerateResetURLTask)(nil)

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
