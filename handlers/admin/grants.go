package admin

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"sort"
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

type GrantFilter struct {
	Section  string
	Item     string
	ItemID   string
	Username string
	RoleName string
	Active   string
	Sort     string
	Dir      string
}

// AdminGrantsAvailablePage lists all available grants.
func AdminGrantsAvailablePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Available Grants"
	data := struct {
		Definitions []*permissions.GrantDefinition
	}{permissions.Definitions}
	AdminGrantsAvailablePageTmpl.Handle(w, r, data)
}

const AdminGrantsAvailablePageTmpl tasks.Template = "admin/grantsAvailablePage.gohtml"

// AdminGrantsPage lists all grants.
func AdminGrantsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Grants"
	queries := cd.Queries()

	filter := GrantFilter{
		Section:  r.URL.Query().Get("section"),
		Item:     r.URL.Query().Get("item"),
		ItemID:   r.URL.Query().Get("item_id"),
		Username: r.URL.Query().Get("user"),
		RoleName: r.URL.Query().Get("role"),
		Active:   r.URL.Query().Get("active"),
		Sort:     r.URL.Query().Get("sort"),
		Dir:      r.URL.Query().Get("dir"),
	}

	params := db.SearchGrantsParams{}
	if filter.Section != "" {
		params.Section = sql.NullString{String: filter.Section, Valid: true}
	}
	if filter.Item != "" {
		params.Item = sql.NullString{String: filter.Item, Valid: true}
	}
	if filter.ItemID != "" {
		if id, err := strconv.Atoi(filter.ItemID); err == nil {
			params.ItemID = sql.NullInt32{Int32: int32(id), Valid: true}
		}
	}
	if filter.Active != "" {
		if b, err := strconv.ParseBool(filter.Active); err == nil {
			params.Active = sql.NullBool{Bool: b, Valid: true}
		}
	}
	if filter.Username != "" {
		params.Username = sql.NullString{String: "%" + filter.Username + "%", Valid: true}
	}
	if filter.RoleName != "" {
		params.RoleName = sql.NullString{String: "%" + filter.RoleName + "%", Valid: true}
	}

	grants, err := queries.SearchGrants(r.Context(), params)
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}
	rows := groupSearchGrants(r.Context(), grants)

	if filter.Sort != "" {
		sortGroups(rows, filter.Sort, filter.Dir)
	}

	data := struct {
		Grants []grantGroup
		Filter GrantFilter
	}{rows, filter}
	AdminGrantsPageTmpl.Handle(w, r, data)
}

const AdminGrantsPageTmpl tasks.Template = "admin/grantsPage.gohtml"

// AdminAnyoneGrantsPage lists grants applying to all users.
func AdminAnyoneGrantsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Grants: Anyone"
	queries := cd.Queries()
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
	data := struct {
		Grants []grantGroup
		Filter GrantFilter
	}{Grants: rows}
	AdminGrantsPageTmpl.Handle(w, r, data)
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
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
	GrantPageTmpl.Handle(w, r, data)
}

const GrantPageTmpl tasks.Template = "admin/grantPage.gohtml"

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

func groupSearchGrants(ctx context.Context, grants []*db.SearchGrantsRow) []grantGroup {
	groups := []*grantGroup{}
	groupMap := map[string]*grantGroup{}
	for _, row := range grants {
		g := &db.Grant{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			UserID:    row.UserID,
			RoleID:    row.RoleID,
			Section:   row.Section,
			Item:      row.Item,
			RuleType:  row.RuleType,
			ItemID:    row.ItemID,
			ItemRule:  row.ItemRule,
			Action:    row.Action,
			Extra:     row.Extra,
			Active:    row.Active,
		}
		key := groupKey(g)
		grp, ok := groupMap[key]
		if !ok {
			grp = &grantGroup{Grant: g, ItemLink: grantItemLink(g)}
			if row.Username.Valid {
				grp.UserName = row.Username.String
			} else if !g.RoleID.Valid {
				grp.UserName = "Anyone"
			}
			if row.RoleName.Valid {
				grp.RoleName = row.RoleName.String
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

func sortGroups(groups []grantGroup, field, dir string) {
	less := func(i, j int) bool {
		a, b := groups[i], groups[j]
		switch field {
		case "user":
			return a.UserName < b.UserName
		case "role":
			return a.RoleName < b.RoleName
		case "section":
			return a.Section < b.Section
		case "item":
			if a.Item.Valid && b.Item.Valid {
				return a.Item.String < b.Item.String
			}
			return a.Item.Valid && !b.Item.Valid
		case "item_id":
			return a.ItemID.Int32 < b.ItemID.Int32
		default:
			return a.ID < b.ID
		}
	}
	if dir == "desc" {
		sort.SliceStable(groups, func(i, j int) bool { return less(j, i) })
	} else {
		sort.SliceStable(groups, less)
	}
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
