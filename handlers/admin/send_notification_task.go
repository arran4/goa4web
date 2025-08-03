package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// SendNotificationTask creates a site notification for users.
type SendNotificationTask struct{ tasks.TaskString }

var sendNotificationTask = &SendNotificationTask{TaskString: TaskNotify}

// ensures SendNotificationTask implements the tasks.Task interface
var _ tasks.Task = (*SendNotificationTask)(nil)
var _ tasks.AuditableTask = (*SendNotificationTask)(nil)

func (SendNotificationTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	message := r.PostFormValue("message")
	link := r.PostFormValue("link")
	role := r.PostFormValue("role")
	names := r.PostFormValue("users")

	var ids []int32
	if names != "" {
		for _, name := range strings.Split(names, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			u, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: name, Valid: true})
			if err != nil {
				return fmt.Errorf("get user %s fail %w", name, handlers.ErrRedirectOnSamePageHandler(err))
			}
			ids = append(ids, u.Idusers)
		}
	} else if role != "" && role != "anonymous" {
		rows, err := queries.AdminListUserIDsByRole(r.Context(), role)
		if err != nil {
			return fmt.Errorf("list role fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		ids = append(ids, rows...)
	} else {
		rows, err := queries.AdminListAllUserIDs(r.Context())
		if err != nil {
			return fmt.Errorf("list users fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		ids = append(ids, rows...)
	}
	for _, id := range ids {
		err := queries.InsertNotification(r.Context(), db.InsertNotificationParams{
			UsersIdusers: id,
			Link:         sql.NullString{String: link, Valid: link != ""},
			Message:      sql.NullString{String: message, Valid: message != ""},
		})
		if err != nil {
			return fmt.Errorf("insert notification fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Count"] = len(ids)
		}
	}
	return nil
}

// AuditRecord summarises sending a site notification.
func (SendNotificationTask) AuditRecord(data map[string]any) string {
	if c, ok := data["Count"].(int); ok {
		return fmt.Sprintf("sent notification to %d users", c)
	}
	return "sent site notification"
}
