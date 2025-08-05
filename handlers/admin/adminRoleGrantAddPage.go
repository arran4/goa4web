package admin

import (
	"fmt"
	"net/http"

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
		Actions     []common.Action
		ItemOptions []ItemOption
	}{CoreData: cd, Role: role, Section: section, Item: item}

	if section == "" {
		for s := range GrantActionMap {
			data.Sections = append(data.Sections, string(s))
		}
	} else if item == "" {
		if items, ok := GrantActionMap[common.Section(section)]; ok {
			for it := range items {
				data.Items = append(data.Items, string(it))
			}
		}
	} else {
		if items, ok := GrantActionMap[common.Section(section)]; ok {
			data.Actions = items[common.Item(item)]
		}
		if common.Section(section) == common.SectionForum && common.Item(item) == common.ItemCategory {
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
