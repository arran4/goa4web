package imagebbs

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// ApprovePostTask marks a post as approved.
type ApprovePostTask struct{ tasks.TaskString }

var _ tasks.Task = (*ApprovePostTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*ApprovePostTask)(nil)
var _ tasks.AuditableTask = (*ApprovePostTask)(nil)

var approvePostTask = &ApprovePostTask{TaskString: TaskApprove}

func (ApprovePostTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	pid, _ := strconv.Atoi(vars["post"])
	if cd == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	boardID, err := imageBoardIDFromRequest(r, cd)
	if err != nil {
		return err
	}
	if !cd.HasGrant("imagebbs", "board", imagebbsApproveAction, boardID) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	queries := cd.Queries()
	if err := queries.AdminApproveImagePost(r.Context(), int32(pid)); err != nil {
		return fmt.Errorf("approve image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["ImagePostID"] = int32(pid)
	}
	return nil
}

func (ApprovePostTask) SelfEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return notif.NewEmailTemplates("imagePostApprovedEmail"), true
}

func (ApprovePostTask) SelfInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := notif.NotificationTemplateFilenameGenerator("image_post_approved")
	return &s
}

func (ApprovePostTask) AuditRecord(data map[string]any) string {
	if id, ok := data["ImagePostID"].(int32); ok {
		return fmt.Sprintf("approved image %d", id)
	}
	return "approved image"
}
