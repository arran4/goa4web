package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// ItemOption represents a selectable item in the add grant form.
type ItemOption struct {
	ID    int32
	Label string
}

// adminRoleGrantAddPage displays a multi-step form for creating a new grant.
func adminRoleGrantAddPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	role, err := cd.SelectedRole()
	if err != nil || role == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("role not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Add Grant: %s", role.Name)

	section := r.URL.Query().Get("section")
	item := r.URL.Query().Get("item")

	data := struct {
		*common.CoreData
		Role        *db.Role
		Section     string
		Item        string
		Sections    []string
		Items       []string
		Actions     []string
		ItemOptions []ItemOption
	}{CoreData: cd, Role: role, Section: section, Item: item}

	if section == "" {
		sectSet := map[string]struct{}{}
		for k := range GrantActionMap {
			parts := strings.Split(k, "|")
			if len(parts) > 0 {
				sectSet[parts[0]] = struct{}{}
			}
		}
		for s := range sectSet {
			data.Sections = append(data.Sections, s)
		}
	} else if item == "" {
		itemSet := map[string]struct{}{}
		for k := range GrantActionMap {
			parts := strings.Split(k, "|")
			if len(parts) == 2 && parts[0] == section {
				itemSet[parts[1]] = struct{}{}
			}
		}
		for it := range itemSet {
			data.Items = append(data.Items, it)
		}
	} else {
		data.Actions = GrantActionMap[section+"|"+item]
		if section == "forum" && item == "category" {
			queries := cd.Queries()
			cats, _ := queries.GetAllForumCategories(r.Context())
			catMap := map[int32]*db.Forumcategory{}
			for _, c := range cats {
				catMap[c.Idforumcategory] = c
			}
			var buildPath func(int32) string
			buildPath = func(id int32) string {
				if id == 0 {
					return ""
				}
				c, ok := catMap[id]
				if !ok || !c.Title.Valid {
					return ""
				}
				parent := buildPath(c.ForumcategoryIdforumcategory)
				if parent != "" {
					return parent + "/" + c.Title.String
				}
				return c.Title.String
			}
			for _, c := range cats {
				label := buildPath(c.Idforumcategory)
				data.ItemOptions = append(data.ItemOptions, ItemOption{ID: c.Idforumcategory, Label: label})
			}
		}
	}

	handlers.TemplateHandler(w, r, "adminRoleGrantAddPage.gohtml", data)
}
