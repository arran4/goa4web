package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

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
	return EmailTemplateAdminNotificationNewsCommentCancel.EmailTemplates(), true
}

func (CancelTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationNewsCommentCancel.NotificationTemplate()
	return &v
}

func (CancelTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateAdminNotificationNewsCommentCancel.RequiredPages()
}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["news"])
	return handlers.RedirectHandler(fmt.Sprintf("/news/news/%d", postId))
}
