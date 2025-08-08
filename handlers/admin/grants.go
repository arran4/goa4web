package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// grantWithNames augments a grant with user and role names.
type grantWithNames struct {
	*db.Grant
	UserName string
	RoleName string
	ItemLink string
}

// AdminGrantsPage lists all grants.
func AdminGrantsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Grants"
	queries := cd.Queries()
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	userNames := map[int32]string{}
	roleNames := map[int32]string{}
	var rows []grantWithNames
	for _, g := range grants {
		gw := grantWithNames{Grant: g}
		if g.UserID.Valid {
			if name, ok := userNames[g.UserID.Int32]; ok {
				gw.UserName = name
			} else if u, err := queries.SystemGetUserByID(r.Context(), g.UserID.Int32); err == nil && u.Username.Valid {
				userNames[g.UserID.Int32] = u.Username.String
				gw.UserName = u.Username.String
			}
		} else if !g.RoleID.Valid {
			gw.UserName = "Anyone"
		}
		if g.RoleID.Valid {
			if name, ok := roleNames[g.RoleID.Int32]; ok {
				gw.RoleName = name
			} else if ro, err := queries.AdminGetRoleByID(r.Context(), g.RoleID.Int32); err == nil {
				roleNames[g.RoleID.Int32] = ro.Name
				gw.RoleName = ro.Name
			}
		}
		gw.ItemLink = grantItemLink(g)
		rows = append(rows, gw)
	}
	data := struct{ Grants []grantWithNames }{rows}
	handlers.TemplateHandler(w, r, "grantsPage.gohtml", data)
}

// AdminAnyoneGrantsPage lists grants applying to all users.
func AdminAnyoneGrantsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Grants: Anyone"
	queries := cd.Queries()
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	var rows []grantWithNames
	for _, g := range grants {
		if g.UserID.Valid || g.RoleID.Valid {
			continue
		}
		gw := grantWithNames{Grant: g, UserName: "Anyone", ItemLink: grantItemLink(g)}
		rows = append(rows, gw)
	}
	data := struct{ Grants []grantWithNames }{rows}
	handlers.TemplateHandler(w, r, "grantsPage.gohtml", data)
}

// adminGrantPage shows a single grant for editing.
func adminGrantPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	idStr := mux.Vars(r)["grant"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("grant not found"))
		return
	}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	var g *db.Grant
	for _, gr := range grants {
		if int(gr.ID) == id {
			g = gr
			break
		}
	}
	if g == nil {
		http.NotFound(w, r)
		return
	}
	gw := grantWithNames{Grant: g}
	if g.UserID.Valid {
		if u, err := queries.SystemGetUserByID(r.Context(), g.UserID.Int32); err == nil && u.Username.Valid {
			gw.UserName = u.Username.String
		}
	} else if !g.RoleID.Valid {
		gw.UserName = "Anyone"
	}
	if g.RoleID.Valid {
		if ro, err := queries.AdminGetRoleByID(r.Context(), g.RoleID.Int32); err == nil {
			gw.RoleName = ro.Name
		}
	}
	gw.ItemLink = grantItemLink(g)
	cd.PageTitle = fmt.Sprintf("Grant %d", g.ID)
	data := struct{ Grant grantWithNames }{gw}
	handlers.TemplateHandler(w, r, "grantPage.gohtml", data)
}

// grantItemLink returns the admin page URL for a grant's item, or "" if none.
func grantItemLink(g *db.Grant) string {
	if !g.ItemID.Valid || g.ItemID.Int32 == 0 {
		return ""
	}
	switch g.Section {
	case "forum":
		switch g.Item.String {
		case "topic":
			return fmt.Sprintf("/admin/forum/topics/topic/%d", g.ItemID.Int32)
		case "category":
			return fmt.Sprintf("/admin/forum/categories/category/%d", g.ItemID.Int32)
		}
	case "linker":
		switch g.Item.String {
		case "category":
			return fmt.Sprintf("/admin/linker/categories/category/%d", g.ItemID.Int32)
		case "link":
			return fmt.Sprintf("/admin/linker/links/link/%d", g.ItemID.Int32)
		}
	case "writing":
		if g.Item.String == "category" {
			return fmt.Sprintf("/admin/writings/categories/category/%d", g.ItemID.Int32)
		}
	}
	return ""
}
