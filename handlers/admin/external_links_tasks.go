package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RefreshExternalLinkTask clears cached preview fields so they will be reloaded.
type RefreshExternalLinkTask struct{ tasks.TaskString }

var refreshExternalLinkTask = &RefreshExternalLinkTask{TaskString: TaskUpdate}

var _ tasks.Task = (*RefreshExternalLinkTask)(nil)
var _ tasks.AuditableTask = (*RefreshExternalLinkTask)(nil)

func (RefreshExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := cd.Queries()
	uid := sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.ClearExternalLinkCache(r.Context(), db.ClearExternalLinkCacheParams{UpdatedBy: uid, AdminID: uid.Int32, ID: int32(id)}); err != nil {
			return fmt.Errorf("clear external link cache fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["RefreshedID"] = appendID(evt.Data["RefreshedID"], id)
		}
	}
	return nil
}

func (RefreshExternalLinkTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["RefreshedID"].(string); ok {
		return "refreshed external links " + ids
	}
	return "refreshed external links"
}

// DeleteExternalLinkTask removes external link entries.
type DeleteExternalLinkTask struct{ tasks.TaskString }

var deleteExternalLinkTask = &DeleteExternalLinkTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteExternalLinkTask)(nil)
var _ tasks.AuditableTask = (*DeleteExternalLinkTask)(nil)

func (DeleteExternalLinkTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := cd.Queries()
	uid := sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeleteExternalLink(r.Context(), db.DeleteExternalLinkParams{ID: int32(id), AdminID: uid.Int32}); err != nil {
			return fmt.Errorf("delete external link fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["DeletedID"] = appendID(evt.Data["DeletedID"], id)
		}
	}
	return nil
}

func (DeleteExternalLinkTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["DeletedID"].(string); ok {
		return "deleted external links " + ids
	}
	return "deleted external links"
}
