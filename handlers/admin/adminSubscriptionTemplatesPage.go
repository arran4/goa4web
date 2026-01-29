package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	subscriptiontemplates "github.com/arran4/goa4web/internal/subscription_templates"
	"github.com/arran4/goa4web/internal/tasks"
)

type subscriptionTemplatePattern struct {
	Method  string
	Pattern string
}

type subscriptionTemplateInfo struct {
	Name     string
	Patterns []subscriptionTemplatePattern
}

// AdminSubscriptionTemplatesPage lists embedded subscription templates and their parsed patterns.
func AdminSubscriptionTemplatesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Subscription Templates"
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		cd.SetCurrentError(errMsg)
	}

	roles, err := cd.Queries().AdminListRoles(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list roles: %w", err))
		return
	}

	names, err := subscriptiontemplates.ListEmbeddedTemplates()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	templates := make([]subscriptionTemplateInfo, 0, len(names))
	for _, name := range names {
		content, err := subscriptiontemplates.GetEmbeddedTemplate(name)
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		parsed := subscriptiontemplates.ParseTemplatePatterns(string(content))
		patterns := make([]subscriptionTemplatePattern, 0, len(parsed))
		for _, entry := range parsed {
			patterns = append(patterns, subscriptionTemplatePattern{Method: entry.Method, Pattern: entry.Pattern})
		}
		templates = append(templates, subscriptionTemplateInfo{Name: name, Patterns: patterns})
	}

	data := struct {
		*common.CoreData
		Templates []subscriptionTemplateInfo
		Roles     []*db.Role
	}{
		CoreData:  cd,
		Templates: templates,
		Roles:     roles,
	}

	AdminSubscriptionTemplatesPageTmpl.Handle(w, r, data)
}

// AdminSubscriptionTemplatesPageTmpl renders the admin subscription templates page.
const AdminSubscriptionTemplatesPageTmpl tasks.Template = "admin/subscriptionTemplatesPage.gohtml"
