package imagebbs

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// ApprovePostTask marks a post as approved.
type ApprovePostTask struct{ tasks.TaskString }

var _ tasks.Task = (*ApprovePostTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*ApprovePostTask)(nil)

var approvePostTask = &ApprovePostTask{TaskString: TaskApprove}

func (ApprovePostTask) Action(w http.ResponseWriter, r *http.Request) any {
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := queries.ApproveImagePost(r.Context(), int32(pid)); err != nil {
		return fmt.Errorf("approve image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

func (ApprovePostTask) SelfEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("imagePostApprovedEmail")
}

func (ApprovePostTask) SelfInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("image_post_approved")
	return &s
}
