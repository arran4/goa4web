package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	subscriptiontemplates "github.com/arran4/goa4web/internal/subscription_templates"
	"github.com/arran4/goa4web/internal/tasks"
)

// ApplySubscriptionTemplateTask applies an embedded subscription template to a role.
type ApplySubscriptionTemplateTask struct {
	tasks.TaskString
	DBPool *sql.DB
}

var _ tasks.Task = (*ApplySubscriptionTemplateTask)(nil)

func (t *ApplySubscriptionTemplateTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	roleName := strings.TrimSpace(r.PostFormValue("role"))
	if roleName == "" {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("role is required"))
	}
	templateName := strings.TrimSpace(r.PostFormValue("template"))
	if templateName == "" {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("template is required"))
	}

	content, err := subscriptiontemplates.GetEmbeddedTemplate(templateName)
	if err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("template %q not found", templateName))
	}

	patterns := subscriptiontemplates.ParseTemplatePatterns(string(content))
	if len(patterns) == 0 {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("template %q has no patterns", templateName))
	}

	seen := make(map[string]struct{}, len(patterns))
	dupes := map[string]struct{}{}
	for _, entry := range patterns {
		key := entry.Method + "\x00" + entry.Pattern
		if _, ok := seen[key]; ok {
			dupes[entry.Method+" "+entry.Pattern] = struct{}{}
			continue
		}
		seen[key] = struct{}{}
	}
	if len(dupes) > 0 {
		entries := make([]string, 0, len(dupes))
		for entry := range dupes {
			entries = append(entries, entry)
		}
		sort.Strings(entries)
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("duplicate patterns: %s", strings.Join(entries, ", ")))
	}

	if t.DBPool == nil {
		return fmt.Errorf("database connection not configured")
	}

	tx, err := t.DBPool.BeginTx(r.Context(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := db.New(tx)
	role, err := qtx.GetRoleByName(r.Context(), roleName)
	if err != nil {
		return handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("unknown role %q", roleName))
	}

	if err := qtx.DeleteSubscriptionArchetypesByRoleAndName(r.Context(), db.DeleteSubscriptionArchetypesByRoleAndNameParams{
		RoleID:        role.ID,
		ArchetypeName: templateName,
	}); err != nil {
		return fmt.Errorf("clean existing archetypes: %w", err)
	}

	for _, entry := range patterns {
		if err := qtx.CreateSubscriptionArchetype(r.Context(), db.CreateSubscriptionArchetypeParams{
			RoleID:        role.ID,
			ArchetypeName: templateName,
			Pattern:       entry.Pattern,
			Method:        entry.Method,
		}); err != nil {
			return fmt.Errorf("insert pattern %s %s: %w", entry.Method, entry.Pattern, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subscription template: %w", err)
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/subscriptions/templates"}
}

// NewApplySubscriptionTemplateTask creates an ApplySubscriptionTemplateTask bound to the admin DB pool.
func (h *Handlers) NewApplySubscriptionTemplateTask() *ApplySubscriptionTemplateTask {
	return &ApplySubscriptionTemplateTask{TaskString: TaskApplySubscriptionTemplate, DBPool: h.DBPool}
}
