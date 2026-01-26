package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// deleteTask removes a queued linker item.
type deleteTask struct{ tasks.TaskString }

var AdminDeleteTask = &deleteTask{TaskString: TaskDelete}
var _ tasks.Task = (*deleteTask)(nil)

var (
	_ tasks.Task                                    = (*deleteTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*deleteTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*deleteTask)(nil)
	_ tasks.EmailTemplatesRequired                  = (*deleteTask)(nil)
)

func (deleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	var link *db.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow
	if rows, err := queries.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails(r.Context()); err == nil {
		for _, it := range rows {
			if it.ID == int32(qid) {
				link = it
				break
			}
		}
	}
	if err := queries.AdminDeleteLinkerQueuedItem(r.Context(), int32(qid)); err != nil {
		return fmt.Errorf("delete linker queued item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if link != nil {
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				u, _ := cd.CurrentUser()
				mod := ""
				if u != nil {
					mod = u.Username.String
				}
				evt.Data["Title"] = link.Title.String
				evt.Data["Username"] = link.Username.String
				evt.Data["Moderator"] = mod
				evt.Data["LinkURL"] = link.Url.String
			}
		}
	}
	return nil
}

func (deleteTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateLinkerRejected.EmailTemplates(), true
}

func (deleteTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateLinkerRejected.NotificationTemplate()
	return &s
}

func (deleteTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationLinkerRejected.EmailTemplates(), true
}

func (deleteTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationLinkerRejected.NotificationTemplate()
	return &v
}

func (deleteTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateLinkerRejected.RequiredPages(), EmailTemplateAdminNotificationLinkerRejected.RequiredPages()...)
}
