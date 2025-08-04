package news

import (
	"fmt"
	"net/http"
	"strconv"

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

func (CancelTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationNewsCommentCancelEmail")
}

func (CancelTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationNewsCommentCancelEmail")
	return &v
}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	postId, _ := strconv.Atoi(vars["news"])
	return handlers.RedirectHandler(fmt.Sprintf("/news/news/%d", postId))
}
