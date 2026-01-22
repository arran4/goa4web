package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserCreateResetLinkTask generates a signed password reset link.
type UserCreateResetLinkTask struct{ tasks.TaskString }

var userCreateResetLinkTask = &UserCreateResetLinkTask{TaskString: TaskCreateResetLink}

var _ tasks.Task = (*UserCreateResetLinkTask)(nil)
var _ tasks.TemplatesRequired = (*UserCreateResetLinkTask)(nil)

const TemplateUserResetLinkPage handlers.Page = "admin/userResetLinkPage.gohtml"

func (UserCreateResetLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("user not found"))
	}

	expiryHours := 24
	if h := r.PostFormValue("expiry_hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 {
			expiryHours = v
		}
	}
	expiry := time.Duration(expiryHours) * time.Hour

	// Create DB entry (no password hash)
	code, err := cd.CreatePasswordResetForUser(user.Idusers, "", "")
	if err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("create reset: %w", err))
	}

	link := cd.SignPasswordResetLink(code, expiry)
	fullLink := cd.AbsoluteURL(link)

	data := struct {
		Link string
		User *db.SystemGetUserByIDRow
		Back string
	}{
		Link: fullLink,
		User: user,
		Back: fmt.Sprintf("/admin/user/%d", user.Idusers),
	}
	return handlers.TemplateWithDataHandler(TemplateUserResetLinkPage, data)
}

func (UserCreateResetLinkTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{TemplateUserResetLinkPage}
}
