package admin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/permissions"
)

// grantWithNames augments a grant with user and role names.
type grantWithNames struct {
	*db.Grant
	UserName string
	RoleName string
	ItemLink string
}

type grantAction struct {
	ID     int32
	Name   string
	Active bool
}

type grantGroup struct {
	*db.Grant
	UserName string
	RoleName string
	ItemLink string
	Actions  []grantAction
}

// AdminGrantsAvailablePage lists all available grants.
func AdminGrantsAvailablePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Available Grants"
	data := struct{ Definitions []permissions.Definition }{permissions.Definitions}
	handlers.TemplateHandler(w, r, "admin/grantsAvailablePage.gohtml", data)
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
	rows := groupGrants(r.Context(), queries, grants)
	data := struct{ Grants []grantGroup }{rows}
	handlers.TemplateHandler(w, r, "admin/grantsPage.gohtml", data)
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
	var filtered []*db.Grant
	for _, g := range grants {
		if g.UserID.Valid || g.RoleID.Valid {
			continue
		}
		filtered = append(filtered, g)
	}
	rows := groupGrants(r.Context(), queries, filtered)
	data := struct{ Grants []grantGroup }{rows}
	handlers.TemplateHandler(w, r, "admin/grantsPage.gohtml", data)
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

func groupGrants(ctx context.Context, queries db.Querier, grants []*db.Grant) []grantGroup {
	userNames := map[int32]string{}
	roleNames := map[int32]string{}
	groups := []*grantGroup{}
	groupMap := map[string]*grantGroup{}
	for _, g := range grants {
		key := groupKey(g)
		grp, ok := groupMap[key]
		if !ok {
			grp = &grantGroup{Grant: g, ItemLink: grantItemLink(g)}
			if g.UserID.Valid {
				if name, ok := userNames[g.UserID.Int32]; ok {
					grp.UserName = name
				} else if u, err := queries.SystemGetUserByID(ctx, g.UserID.Int32); err == nil && u.Username.Valid {
					userNames[g.UserID.Int32] = u.Username.String
					grp.UserName = u.Username.String
				}
			} else if !g.RoleID.Valid {
				grp.UserName = "Anyone"
			}
			if g.RoleID.Valid {
				if name, ok := roleNames[g.RoleID.Int32]; ok {
					grp.RoleName = name
				} else if ro, err := queries.AdminGetRoleByID(ctx, g.RoleID.Int32); err == nil {
					roleNames[g.RoleID.Int32] = ro.Name
					grp.RoleName = ro.Name
				}
			}
			groupMap[key] = grp
			groups = append(groups, grp)
		}
		grp.Actions = append(grp.Actions, grantAction{ID: g.ID, Name: g.Action, Active: g.Active})
	}
	res := make([]grantGroup, len(groups))
	for i, g := range groups {
		res[i] = *g
	}
	return res
}

func groupKey(g *db.Grant) string {
	user := ""
	if g.UserID.Valid {
		user = strconv.Itoa(int(g.UserID.Int32))
	}
	role := ""
	if g.RoleID.Valid {
		role = strconv.Itoa(int(g.RoleID.Int32))
	}
	item := ""
	if g.Item.Valid {
		item = g.Item.String
	}
	itemID := ""
	if g.ItemID.Valid {
		itemID = strconv.Itoa(int(g.ItemID.Int32))
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s", user, role, g.Section, item, itemID)
}
