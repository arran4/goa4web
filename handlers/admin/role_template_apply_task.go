package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	roletemplates "github.com/arran4/goa4web/internal/role_templates"
	"github.com/arran4/goa4web/internal/tasks"
)

// RoleTemplateApplyTask applies a role template to the database.
type RoleTemplateApplyTask struct{ tasks.TaskString }

var roleTemplateApplyTask = &RoleTemplateApplyTask{TaskString: TaskApplyRoleTemplate}

var _ tasks.Task = (*RoleTemplateApplyTask)(nil)
var _ tasks.AuditableTask = (*RoleTemplateApplyTask)(nil)

func (RoleTemplateApplyTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { handlers.RenderErrorPage(w, r, handlers.ErrForbidden) })
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	name := r.PostFormValue("name")
	if name == "" {
		return fmt.Errorf("template name is required %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("template name is required")))
	}
	tmpl, ok := roletemplates.Templates[name]
	if !ok {
		return fmt.Errorf("template %q not found %w", name, handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("template %q not found", name)))
	}

	queries, ok := cd.Queries().(*db.Queries)
	if !ok {
		return fmt.Errorf("role template apply requires sqlc queries %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("invalid queries implementation")))
	}

	tx, err := queries.BeginTx(r.Context(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)
	if err := roletemplates.ApplyRoles(r.Context(), qtx, tx, tmpl.Roles, time.Now(), nil); err != nil {
		return fmt.Errorf("apply template: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["TemplateName"] = name
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/roles/templates?name=" + url.QueryEscape(name)}
}

func (RoleTemplateApplyTask) AuditRecord(data map[string]any) string {
	if name, ok := data["TemplateName"].(string); ok {
		return fmt.Sprintf("applied role template %s", name)
	}
	return "applied role template"
}
