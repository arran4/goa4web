package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteTemplateTask removes a template override.
type DeleteTemplateTask struct{ tasks.TaskString }

var deleteTemplateTask = &DeleteTemplateTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteTemplateTask)(nil)
var _ tasks.AuditableTask = (*DeleteTemplateTask)(nil)

func (DeleteTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	name := r.PostFormValue("name")
	q := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := q.AdminDeleteTemplateOverride(r.Context(), name); err != nil {
		return fmt.Errorf("db delete template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Template"] = name
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/email/template?name=" + name}
}

func (DeleteTemplateTask) AuditRecord(data map[string]any) string {
	if t, ok := data["Template"].(string); ok {
		return "deleted template override " + t
	}
	return "deleted template override"
}
