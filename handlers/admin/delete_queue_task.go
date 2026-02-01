package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteQueueTask removes queued emails without sending.
type DeleteQueueTask struct{ tasks.TaskString }

var deleteQueueTask = &DeleteQueueTask{TaskString: TaskDelete}

// ensure DeleteQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*DeleteQueueTask)(nil)
var _ tasks.AuditableTask = (*DeleteQueueTask)(nil)

func (DeleteQueueTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	var ids []int32
	if r.Form.Get("selection") == "filtered" {
		filters := emailFiltersFromRequest(r)
		if strings.Contains(r.URL.Path, "/failed") {
			rows, err := queries.AdminListFailedEmailIDs(r.Context(), db.AdminListFailedEmailIDsParams{
				LanguageID:    filters.LangIDParam(),
				RoleName:      filters.RoleParam(),
				Provider:      filters.ProviderParam(),
				CreatedBefore: filters.CreatedBefore,
			})
			if err != nil {
				return fmt.Errorf("list failed emails fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			for _, id := range rows {
				ids = append(ids, id)
			}
		} else {
			rows, err := queries.AdminListUnsentPendingEmails(r.Context(), db.AdminListUnsentPendingEmailsParams{
				LanguageID:    filters.LangIDParam(),
			RoleName:      filters.RoleParam(),
				Status:        filters.StatusParam(),
				Provider:      filters.ProviderParam(),
				CreatedBefore: filters.CreatedBefore,
				Limit:         2147483647,
				Offset:        0,
			})
			if err != nil {
				return fmt.Errorf("list pending emails fail %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			for _, row := range rows {
				ids = append(ids, row.ID)
			}
		}
	} else {
		for _, idStr := range r.Form["id"] {
			id, _ := strconv.Atoi(idStr)
			ids = append(ids, int32(id))
		}
	}

	for _, id := range ids {
		if err := queries.AdminDeletePendingEmail(r.Context(), int32(id)); err != nil {
			return fmt.Errorf("delete email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["DeletedEmailID"] = appendID(evt.Data["DeletedEmailID"], int(id))
			}
		}
	}
	return nil
}

// AuditRecord summarises queued emails being removed.
func (DeleteQueueTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["DeletedEmailID"].(string); ok {
		return "deleted queued emails " + ids
	}
	return "deleted queued emails"
}
