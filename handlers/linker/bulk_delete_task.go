package linker

import (
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// bulkDeleteTask removes multiple queued linker items.
type bulkDeleteTask struct{ tasks.TaskString }

var BulkDeleteTask = &bulkDeleteTask{TaskString: TaskBulkDelete}

var (
	_ tasks.Task                                    = (*bulkDeleteTask)(nil)
	_ notif.SubscribersNotificationTemplateProvider = (*bulkDeleteTask)(nil)
	_ notif.AdminEmailTemplateProvider              = (*bulkDeleteTask)(nil)
)

func (bulkDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
	}
	var info []map[string]any
	if rows, err := queries.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails(r.Context()); err == nil {
		ids := make(map[int]struct{})
		for _, q := range r.Form["qid"] {
			id, _ := strconv.Atoi(q)
			ids[id] = struct{}{}
		}
		for _, it := range rows {
			if _, ok := ids[int(it.Idlinkerqueue)]; ok {
				info = append(info, map[string]any{"Title": it.Title.String, "URL": it.Url.String, "Username": it.Username.String})
			}
		}
	}
	for _, q := range r.Form["qid"] {
		id, _ := strconv.Atoi(q)
		if err := queries.DeleteLinkerQueuedItem(r.Context(), int32(id)); err != nil {
			log.Printf("deleteLinkerQueuedItem Error: %s", err)
		}
	}
	if len(info) > 0 {
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
				for i := range info {
					info[i]["Moderator"] = mod
				}
				evt.Data["links"] = info
				if len(info) == 1 {
					if url, ok := info[0]["URL"].(string); ok {
						evt.Data["LinkURL"] = url
					}
				}
				evt.Data["Moderator"] = mod
			}
		}
	}
	return nil
}

func (bulkDeleteTask) SubscribedEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("linkerRejectedEmail")
}

func (bulkDeleteTask) SubscribedInternalNotificationTemplate() *string {
	s := notif.NotificationTemplateFilenameGenerator("linker_rejected")
	return &s
}

func (bulkDeleteTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationLinkerRejectedEmail")
}

func (bulkDeleteTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationLinkerRejectedEmail")
	return &v
}
