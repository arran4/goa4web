package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// GrantDefinition describes the allowed actions for a grant and any additional
// rules that apply.
type GrantDefinition struct {
	Actions       []string
	RequireItemID bool // whether grants must specify an item_id
}

// GrantActionMap defines allowed actions and rules for each section and item
// combination. Key format: "section|item". Keep in sync with
// specs/permissions.md.
var GrantActionMap = map[string]GrantDefinition{
	"forum|":              {Actions: []string{"search"}},
	"forum|topic":         {Actions: []string{"see", "view", "reply", "post", "edit"}, RequireItemID: true},
	"forum|thread":        {Actions: []string{"see", "view", "reply", "edit"}, RequireItemID: true},
	"forum|comment":       {Actions: []string{"edit"}},
	"forum|category":      {Actions: []string{"see", "view", "post"}, RequireItemID: true},
	"linker|":             {Actions: []string{"search"}},
	"linker|category":     {Actions: []string{"see", "view", "post"}},
	"linker|link":         {Actions: []string{"see", "view", "reply"}},
	"imagebbs|":           {Actions: []string{"search"}},
	"imagebbs|board":      {Actions: []string{"see", "view", "post", "approve"}},
	"images|upload":       {Actions: []string{"see", "post"}},
	"news|":               {Actions: []string{"search"}},
	"news|post":           {Actions: []string{"see", "view", "reply", "post", "edit"}},
	"blogs|":              {Actions: []string{"search"}},
	"blogs|entry":         {Actions: []string{"see", "view", "reply", "post", "edit"}},
	"writing|":            {Actions: []string{"search"}},
	"writing|category":    {Actions: []string{"see", "view", "post"}},
	"writing|article":     {Actions: []string{"see", "view", "reply", "post", "edit"}},
	"faq|":                {Actions: []string{"search"}},
	"faq|category":        {Actions: []string{"see", "view"}},
	"faq|question":        {Actions: []string{"post"}},
	"faq|question/answer": {Actions: []string{"see", "view"}},
	"search|":             {Actions: []string{"search"}},
	"privateforum|topic":  {Actions: []string{"see", "view", "reply", "post", "edit", "create"}},
}

// GrantAction represents a single grant action and whether it's unsupported.
type GrantAction struct {
	Name        string
	Unsupported bool
}

// GrantGroup represents grants grouped by section and item for editing.
type GrantGroup struct {
	Section     string
	Item        string
	ItemID      sql.NullInt32
	Link        string
	Info        string
	Have        []GrantAction
	Disabled    []GrantAction
	Available   []string
	Unsupported bool
}

// buildGrantGroups loads grants for a role and organises them for the grants editor.
func buildGrantGroups(ctx context.Context, cd *common.CoreData, roleID int32) ([]GrantGroup, error) {
	queries := cd.Queries()

	grants, err := queries.AdminListGrantsByRoleID(ctx, sql.NullInt32{Int32: roleID, Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return buildGrantGroupsFromGrants(ctx, cd, grants)
}

// BuildGrantGroups loads grants for a role and organises them for export.
func BuildGrantGroups(ctx context.Context, cd *common.CoreData, roleID int32) ([]GrantGroup, error) {
	return buildGrantGroups(ctx, cd, roleID)
}

// buildGrantGroupsForUser loads grants for a user and organises them for the grants editor.
func buildGrantGroupsForUser(ctx context.Context, cd *common.CoreData, userID int32) ([]GrantGroup, error) {
	queries := cd.Queries()

	grants, err := queries.ListGrantsByUserID(ctx, sql.NullInt32{Int32: userID, Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return buildGrantGroupsFromGrants(ctx, cd, grants)
}

// buildGrantGroupsFromGrants organises grants for the grants editor.
func buildGrantGroupsFromGrants(ctx context.Context, cd *common.CoreData, grants []*db.Grant) ([]GrantGroup, error) {
	queries := cd.Queries()

	forumCats, _ := queries.GetAllForumCategories(ctx, db.GetAllForumCategoriesParams{ViewerID: 0})
	catMap := map[int32]*db.Forumcategory{}
	for _, c := range forumCats {
		catMap[c.Idforumcategory] = c
	}

	langs, _ := cd.Languages()
	langMap := map[int32]string{}
	for _, l := range langs {
		if l.Nameof.Valid {
			langMap[l.ID] = l.Nameof.String
		}
	}

	buildCatPath := func(id int32) string {
		if id == 0 {
			return ""
		}
		var parts []string
		for cid := id; cid != 0; {
			c, ok := catMap[cid]
			if !ok || !c.Title.Valid {
				break
			}
			parts = append([]string{c.Title.String}, parts...)
			cid = c.ForumcategoryIdforumcategory
		}
		return strings.Join(parts, "/")
	}

	type GrantInfo struct {
		*db.Grant
		Link string
		Info string
	}
	var ginfos []GrantInfo
	for _, g := range grants {
		if def, ok := GrantActionMap[g.Section+"|"+g.Item.String]; ok && def.RequireItemID && (!g.ItemID.Valid || g.ItemID.Int32 == 0) {
			continue
		}
		gi := GrantInfo{Grant: g}
		if g.Item.Valid && g.ItemID.Valid {
			switch g.Section {
			case "forum":
				switch g.Item.String {
				case "topic":
					gi.Link = fmt.Sprintf("/admin/forum/topics/topic/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if t, err := queries.GetForumTopicById(ctx, g.ItemID.Int32); err == nil {
						if t.Title.Valid {
							info := t.Title.String
							cat := buildCatPath(t.ForumcategoryIdforumcategory)
							if cat != "" {
								info = fmt.Sprintf("%s (%s)", info, cat)
							}
							gi.Info = info
						}
					}
				case "category":
					gi.Link = fmt.Sprintf("/admin/forum/categories/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if c, err := queries.GetForumCategoryById(ctx, db.GetForumCategoryByIdParams{Idforumcategory: g.ItemID.Int32, ViewerID: 0}); err == nil && c.Title.Valid {
						path := buildCatPath(c.Idforumcategory)
						gi.Info = path
					}
				case "thread":
					if tid, err := queries.GetForumTopicIdByThreadId(ctx, g.ItemID.Int32); err == nil {
						if t, err := queries.GetForumTopicById(ctx, tid); err == nil {
							if t.Title.Valid {
								cat := buildCatPath(t.ForumcategoryIdforumcategory)
								info := fmt.Sprintf("%s thread", t.Title.String)
								if cat != "" {
									info = fmt.Sprintf("%s (%s)", info, cat)
								}
								gi.Info = info
							}
						}
					}
				}
			case "linker":
				switch g.Item.String {
				case "category":
					gi.Link = fmt.Sprintf("/admin/linker/categories/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if c, err := queries.GetLinkerCategoryById(ctx, g.ItemID.Int32); err == nil && c.Title.Valid {
						gi.Info = c.Title.String
					}
				case "link":
					gi.Link = fmt.Sprintf("/admin/linker/links/link/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if l, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(ctx, g.ItemID.Int32); err == nil && l.Title.Valid {
						gi.Info = l.Title.String
					}
				}
			case "writing":
				switch g.Item.String {
				case "category":
					gi.Link = fmt.Sprintf("/admin/writings/categories/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if c, err := queries.GetWritingCategoryById(ctx, g.ItemID.Int32); err == nil && c.Title.Valid {
						gi.Info = c.Title.String
					}
				case "article":
					if w, err := queries.GetWritingForListerByID(ctx, db.GetWritingForListerByIDParams{ListerID: cd.UserID, Idwriting: g.ItemID.Int32, ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}}); err == nil {
						if w.Title.Valid {
							info := w.Title.String
							if name, ok := langMap[w.LanguageID.Int32]; ok && name != "" {
								info = fmt.Sprintf("[%s] %s", name, info)
							}
							gi.Info = info
						}
					}
				}
			case "faq":
				switch g.Item.String {
				case "category":
					if cats, err := queries.AdminGetFAQCategories(ctx); err == nil {
						for _, c := range cats {
							if c.ID == g.ItemID.Int32 {
								if c.Name.Valid {
									gi.Info = c.Name.String
								}
								break
							}
						}
					}
				case "question", "question/answer":
					if qrow, err := queries.AdminGetFAQByID(ctx, g.ItemID.Int32); err == nil && qrow.Question.Valid {
						text := qrow.Question.String
						if len(text) > 40 {
							text = text[:40] + "..."
						}
						if qrow.LanguageID.Valid {
							if name, ok := langMap[qrow.LanguageID.Int32]; ok && name != "" {
								text = fmt.Sprintf("[%s] %s", name, text)
							}
						}
						gi.Info = text
					}
				}
			case "imagebbs":
				if g.Item.String == "board" {
					if b, err := queries.GetImageBoardById(ctx, g.ItemID.Int32); err == nil && b.Title.Valid {
						gi.Info = b.Title.String
					}
				}
			}
		} else if g.Section == "role" && g.Action != "" {
			if roles, err := cd.AllRoles(); err == nil {
				for _, ro := range roles {
					if ro.Name == g.Action {
						gi.Link = fmt.Sprintf("/admin/role/%d#g%d", ro.ID, g.ID)
						break
					}
				}
			}
		}
		ginfos = append(ginfos, gi)
	}

	groupMap := map[string]*GrantGroup{}
	for _, gi := range ginfos {
		key := fmt.Sprintf("%s|%s|%d", gi.Section, gi.Item.String, gi.ItemID.Int32)
		grp, ok := groupMap[key]
		if !ok {
			grp = &GrantGroup{Section: gi.Section, Item: gi.Item.String, ItemID: gi.ItemID, Link: gi.Link, Info: gi.Info}
			if _, ok := GrantActionMap[gi.Section+"|"+gi.Item.String]; !ok {
				grp.Unsupported = true
			}
			groupMap[key] = grp
		}
		ga := GrantAction{Name: gi.Action}
		if def, ok := GrantActionMap[gi.Section+"|"+gi.Item.String]; ok {
			actSet := map[string]struct{}{}
			for _, a := range def.Actions {
				actSet[a] = struct{}{}
			}
			if _, ok := actSet[gi.Action]; !ok {
				ga.Unsupported = true
			}
		} else {
			ga.Unsupported = true
		}
		if gi.Active {
			grp.Have = append(grp.Have, ga)
		} else {
			grp.Disabled = append(grp.Disabled, ga)
		}
	}

	// Ensure all section/item pairs appear even when the role has no grants.
	for key, def := range GrantActionMap {
		if def.RequireItemID {
			continue
		}
		parts := strings.Split(key, "|")
		if len(parts) != 2 {
			continue
		}
		gkey := fmt.Sprintf("%s|%s|0", parts[0], parts[1])
		if _, ok := groupMap[gkey]; !ok {
			groupMap[gkey] = &GrantGroup{Section: parts[0], Item: parts[1], ItemID: sql.NullInt32{}}
		}
	}

	groups := make([]GrantGroup, 0, len(groupMap))
	for _, grp := range groupMap {
		if def, ok := GrantActionMap[grp.Section+"|"+grp.Item]; ok {
			haveSet := map[string]struct{}{}
			for _, h := range grp.Have {
				haveSet[h.Name] = struct{}{}
			}
			for _, d := range grp.Disabled {
				haveSet[d.Name] = struct{}{}
			}
			for _, a := range def.Actions {
				if _, ok := haveSet[a]; !ok {
					grp.Available = append(grp.Available, a)
				}
			}
		} else {
			grp.Unsupported = true
		}
		groups = append(groups, *grp)
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Section != groups[j].Section {
			return groups[i].Section < groups[j].Section
		}
		if groups[i].Item != groups[j].Item {
			return groups[i].Item < groups[j].Item
		}
		return groups[i].ItemID.Int32 < groups[j].ItemID.Int32
	})

	return groups, nil
}
