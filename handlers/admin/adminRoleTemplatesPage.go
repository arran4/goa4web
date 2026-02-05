package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	roletemplates "github.com/arran4/goa4web/internal/role_templates"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminRoleTemplatesPage lists role templates and shows diffs against current roles.
func AdminRoleTemplatesPage(w http.ResponseWriter, r *http.Request) {
	type TemplateListItem struct {
		Name        string
		Description string
		Active      bool
	}
	type RolePreview struct {
		Role        roletemplates.RoleDef
		GrantGroups []GrantGroup
	}
	type DiffSummary struct {
		NewRoles      int
		UpdatedRoles  int
		MatchingRoles int
		GrantsAdded   int
		GrantsRemoved int
	}
	type TemplateDetail struct {
		Template     roletemplates.TemplateDef
		Diffs        []roletemplates.RoleDiff
		Summary      DiffSummary
		RolePreviews []RolePreview
	}
	type Data struct {
		Templates []TemplateListItem
		Selected  *TemplateDetail
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Role Templates"

	names := roletemplates.SortedTemplateNames()
	items := make([]TemplateListItem, 0, len(names))
	selectedName := r.URL.Query().Get("name")
	if selectedName == "" && len(names) > 0 {
		selectedName = names[0]
	}

	for _, name := range names {
		tmpl := roletemplates.Templates[name]
		items = append(items, TemplateListItem{
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Active:      name == selectedName,
		})
	}

	var detail *TemplateDetail
	if selectedName != "" {
		tmpl, ok := roletemplates.Templates[selectedName]
		if !ok {
			log.Printf("role template %q not found", selectedName)
			handlers.RenderErrorPage(w, r, fmt.Errorf("template %q not found", selectedName))
			return
		}
		diffs, err := roletemplates.BuildTemplateDiff(r.Context(), cd.Queries(), tmpl)
		if err != nil {
			log.Printf("build role template diff: %v", err)
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}

		summary := DiffSummary{}
		for _, diff := range diffs {
			if diff.Status == "new" {
				summary.NewRoles++
			} else {
				if len(diff.PropertyChanges) > 0 || len(diff.GrantsAdded) > 0 || len(diff.GrantsRemoved) > 0 {
					summary.UpdatedRoles++
				} else {
					summary.MatchingRoles++
				}
			}
			summary.GrantsAdded += len(diff.GrantsAdded)
			summary.GrantsRemoved += len(diff.GrantsRemoved)
		}

		rolePreviews := make([]RolePreview, 0, len(tmpl.Roles))
		for _, role := range tmpl.Roles {
			grants := templateGrantsToDB(role.Grants)
			groups, err := buildGrantGroupsFromGrants(r.Context(), cd, grants)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("build role template grant preview: %v", err)
				handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
				return
			}
			rolePreviews = append(rolePreviews, RolePreview{Role: role, GrantGroups: groups})
		}
		detail = &TemplateDetail{
			Template:     tmpl,
			Diffs:        diffs,
			Summary:      summary,
			RolePreviews: rolePreviews,
		}
	}

	data := Data{
		Templates: items,
		Selected:  detail,
	}
	AdminRoleTemplatesPageTmpl.Handle(w, r, data)
}

func templateGrantsToDB(grants []roletemplates.GrantDef) []*db.Grant {
	out := make([]*db.Grant, 0, len(grants))
	for i, g := range grants {
		out = append(out, &db.Grant{
			ID:      int32(i + 1),
			Section: g.Section,
			Item:    sql.NullString{String: g.Item, Valid: g.Item != ""},
			Action:  g.Action,
			ItemID:  sql.NullInt32{Int32: g.ItemID, Valid: g.ItemID != 0},
			Active:  true,
		})
	}
	return out
}

// AdminRoleTemplatesPageTmpl renders the role templates admin page.
const AdminRoleTemplatesPageTmpl tasks.Template = "admin/roleTemplatesPage.gohtml"
