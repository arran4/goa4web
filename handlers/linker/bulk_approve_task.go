package linker

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/eventbus"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"
)

// bulkApproveTask approves multiple queued linker items.
type bulkApproveTask struct{ tasks.TaskString }

var AdminBulkApproveTask = &bulkApproveTask{TaskString: TaskBulkApprove}

var (
	_ tasks.Task                                    = (*bulkApproveTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*bulkApproveTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*bulkApproveTask)(nil)
	_ tasks.EmailTemplatesRequired                  = (*bulkApproveTask)(nil)
	_ searchworker.IndexedTask                      = bulkApproveTask{}
)

func (bulkApproveTask) IndexType() string { return searchworker.TypeLinker }

func (bulkApproveTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

func (bulkApproveTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
	}
	var links []map[string]any
	for _, q := range r.Form["qid"] {
		id, _ := strconv.Atoi(q)
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		lid, err := queries.AdminInsertQueuedLinkFromQueue(r.Context(), int32(id))
		if err != nil {
			log.Printf("selectInsert Error: %s", err)
			continue
		}
		link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser(r.Context(), db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUserParams{
			ViewerID:     cd.UserID,
			ID:           int32(lid),
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			log.Printf("getLinkerItemById Error: %s", err)
			continue
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
				links = append(links, map[string]any{"Title": link.Title.String, "URL": link.Url.String, "Username": link.Username.String, "Moderator": mod})
				evt.Data["Moderator"] = mod
				evt.Data["links"] = links
				if len(links) == 1 {
					evt.Data["LinkURL"] = cd.AbsoluteURL(fmt.Sprintf("/linker/show/%d", lid))
				}
				evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeLinker, ID: int32(lid), Text: text}
			}
		}
	}
	return nil
}

func (bulkApproveTask) SubscribedEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateLinkerApproved.EmailTemplates(), true
}

func (bulkApproveTask) SubscribedInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	s := NotificationTemplateLinkerApproved.NotificationTemplate()
	return &s
}

func (bulkApproveTask) AdminEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateAdminNotificationLinkerApproved.EmailTemplates(), true
}

func (bulkApproveTask) AdminInternalNotificationTemplate(evt eventbus.TaskEvent) *string {
	v := EmailTemplateAdminNotificationLinkerApproved.NotificationTemplate()
	return &v
}

func (bulkApproveTask) EmailTemplatesRequired() []tasks.Page {
	return append(EmailTemplateLinkerApproved.RequiredPages(), EmailTemplateAdminNotificationLinkerApproved.RequiredPages()...)
}
