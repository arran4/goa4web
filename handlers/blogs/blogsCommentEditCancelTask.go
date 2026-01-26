package blogs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

var _ tasks.Task = (*CancelTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*CancelTask)(nil)
var _ tasks.EmailTemplatesRequired = (*CancelTask)(nil)

func (CancelTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationBlogCommentCancel.EmailTemplates(), true
}

func (CancelTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationBlogCommentCancel.NotificationTemplate()
	return &v
}

func (CancelTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationBlogCommentCancel.RequiredPages()
}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
	return handlers.RedirectHandler(fmt.Sprintf("/blogs/blog/%d/comments", blogId))
}
