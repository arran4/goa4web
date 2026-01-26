package linker

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

// approveTask approves a queued linker item.
type approveTask struct{ tasks.TaskString }

var AdminApproveTask = &approveTask{TaskString: TaskApprove}
var _ tasks.Task = (*approveTask)(nil)

var (
	_ tasks.Task                                    = (*approveTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*approveTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*approveTask)(nil)
	_ tasks.EmailTemplatesRequired                  = (*approveTask)(nil)
	_ searchworker.IndexedTask                      = approveTask{}
)

func (approveTask) IndexType() string { return searchworker.TypeLinker }

func (approveTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (approveTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	lid, err := queries.AdminInsertQueuedLinkFromQueue(r.Context(), int32(qid))
	if err != nil {
		return fmt.Errorf("approve linker item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
		ViewerID:     cd.UserID,
		ID:           int32(lid),
		ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		return fmt.Errorf("get linker item fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	text := strings.Join([]string{link.Title.String, link.Description.String}, " ")
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
			evt.Data["LinkURL"] = cd.AbsoluteURL(fmt.Sprintf("/linker/show/%d", lid))
			evt.Data["UnsubURL"] = cd.AbsoluteURL("/usr/subscriptions")
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeLinker, ID: int32(lid), Text: text}
		}
	}
	return nil
}

func (approveTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateLinkerApproved.EmailTemplates(), true
}

func (approveTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateLinkerApproved.NotificationTemplate()
	return &s
}

func (approveTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationLinkerApproved.EmailTemplates(), true
}

func (approveTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationLinkerApproved.NotificationTemplate()
	return &v
}

func (approveTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateLinkerApproved.RequiredPages(), EmailTemplateAdminNotificationLinkerApproved.RequiredPages()...)
}
