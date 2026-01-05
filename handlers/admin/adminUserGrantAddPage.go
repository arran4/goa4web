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

// adminUserGrantAddPage displays a multi-step form for creating a new grant for a user.
func adminUserGrantAddPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	user := cd.CurrentProfileUser()
	if user == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	cd.PageTitle = fmt.Sprintf("Add Grant to User: %s", user.Username.String)

	section := r.URL.Query().Get("section")
	item := r.URL.Query().Get("item")

	data := struct {
		User          *db.SystemGetUserByIDRow
		Section       string
		Item          string
		Sections      []string
		Items         []string
		Actions       []string
		ItemOptions   []ItemOption
		RequireItemID bool
	}{User: user, Section: section, Item: item}

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
	} else {
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
		if item == "" && len(data.Items) > 0 {
			item = data.Items[0]
			data.Item = item
		}
		def := GrantActionMap[section+"|"+item]
		data.Actions = def.Actions
		data.RequireItemID = def.RequireItemID
		if section == "forum" && item == "category" {
			queries := cd.Queries()
			cats, _ := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
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

	handlers.TemplateHandler(w, r, TemplateUserGrantAddPage, data)
}
