package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"
)

// SaveTemplateTask stores a custom update email template.
type SaveTemplateTask struct{ tasks.TaskString }

var saveTemplateTask = &SaveTemplateTask{TaskString: TaskUpdate}

// compile-time interface check for SaveTemplateTask
var _ tasks.Task = (*SaveTemplateTask)(nil)
var _ tasks.AuditableTask = (*SaveTemplateTask)(nil)

func (SaveTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	body := r.PostFormValue("body")
	q := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := q.SetTemplateOverride(r.Context(), db.SetTemplateOverrideParams{Name: "updateEmail", Body: body}); err != nil {
		return fmt.Errorf("db save template fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Template"] = "updateEmail"
		}
	}
	return handlers.RedirectHandler("/admin/email/template")
}

// AuditRecord summarises saving the update email template.
func (SaveTemplateTask) AuditRecord(data map[string]any) string {
	if t, ok := data["Template"].(string); ok {
		return "saved template " + t
	}
	return "saved email template"
}
