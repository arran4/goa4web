package admin

import (
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers/admincommon"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminGrantAddPage struct{}

func (p *AdminGrantAddPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Add Grant"
	queries := cd.Queries()

	userInfo, _ := admincommon.LoadUserRoleInfo(r.Context(), queries, nil)
	roles, _ := queries.AdminListRoles(r.Context())

	definitions := map[string]map[string]GrantDefinition{}
	for key, def := range GrantActionMap {
		parts := strings.Split(key, "|")
		section := parts[0]
		item := ""
		if len(parts) == 2 {
			item = parts[1]
		}
		if _, ok := definitions[section]; !ok {
			definitions[section] = map[string]GrantDefinition{}
		}
		definitions[section][item] = def
	}

	itemOptions := map[string][]ItemOption{}
	cats, _ := queries.GetAllForumCategories(r.Context(), db.GetAllForumCategoriesParams{ViewerID: 0})
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
		itemOptions["forum|category"] = append(itemOptions["forum|category"], ItemOption{ID: c.Idforumcategory, Label: label})
	}

	data := struct {
		Users           []admincommon.UserRoleInfo
		Roles           []*db.Role
		GrantActions    map[string]map[string]GrantDefinition
		GrantItemLookup map[string][]ItemOption
	}{
		Users:           userInfo,
		Roles:           roles,
		GrantActions:    definitions,
		GrantItemLookup: itemOptions,
	}

	AdminGrantAddPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminGrantAddPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Add Grant", "/admin/grant/add", &AdminPage{}
}

func (p *AdminGrantAddPage) PageTitle() string {
	return "Add Grant"
}

var _ common.Page = (*AdminGrantAddPage)(nil)
var _ http.Handler = (*AdminGrantAddPage)(nil)

const AdminGrantAddPageTmpl tasks.Template = "admin/grantAddPage.gohtml"
