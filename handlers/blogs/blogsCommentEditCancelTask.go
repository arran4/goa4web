package blogs

import (
	"fmt"
	"net/http"
	"strconv"

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

func (CancelTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationBlogCommentCancelEmail")
}

func (CancelTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationBlogCommentCancelEmail")
	return &v
}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)

	vars := mux.Vars(r)
	blogId, _ := strconv.Atoi(vars["blog"])
	return handlers.RedirectHandler(fmt.Sprintf("/blogs/blog/%d/comments", blogId))
}
