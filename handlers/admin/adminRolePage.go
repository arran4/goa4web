package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// adminRolePage shows details for a role including grants and users.
func adminRolePage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	role, err := queries.AdminGetRoleByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}
	cd.PageTitle = fmt.Sprintf("Role %s", role.Name)

	users, err := queries.AdminListUsersByRoleID(r.Context(), int32(id))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	grants, err := queries.AdminListGrantsByRoleID(r.Context(), sql.NullInt32{Int32: int32(id), Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Preload forum categories and languages for context information.
	forumCats, _ := queries.GetAllForumCategories(r.Context())
	catMap := map[int32]*db.Forumcategory{}
	for _, c := range forumCats {
		catMap[c.Idforumcategory] = c
	}
	langs, _ := cd.Languages()
	langMap := map[int32]string{}
	for _, l := range langs {
		if l.Nameof.Valid {
			langMap[l.Idlanguage] = l.Nameof.String
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
		gi := GrantInfo{Grant: g}
		if g.Item.Valid && g.ItemID.Valid {
			switch g.Section {
			case "forum":
				switch g.Item.String {
				case "topic":
					gi.Link = fmt.Sprintf("/admin/forum/topic/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if t, err := queries.GetForumTopicById(r.Context(), g.ItemID.Int32); err == nil {
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
					gi.Link = fmt.Sprintf("/admin/forum/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if c, err := queries.GetForumCategoryById(r.Context(), g.ItemID.Int32); err == nil && c.Title.Valid {
						path := buildCatPath(c.Idforumcategory)
						gi.Info = path
					}
				case "thread":
					if tid, err := queries.GetForumTopicIdByThreadId(r.Context(), g.ItemID.Int32); err == nil {
						if t, err := queries.GetForumTopicById(r.Context(), tid); err == nil {
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
					gi.Link = fmt.Sprintf("/admin/linker/category/%d/grants#g%d", g.ItemID.Int32, g.ID)
					if c, err := queries.GetLinkerCategoryById(r.Context(), g.ItemID.Int32); err == nil && c.Title.Valid {
						gi.Info = c.Title.String
					}
				case "link":
					if l, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), g.ItemID.Int32); err == nil && l.Title.Valid {
						gi.Info = l.Title.String
					}
				}
			case "writings":
				switch g.Item.String {
				case "category":
					gi.Link = fmt.Sprintf("/admin/writings/category/%d/permissions#g%d", g.ItemID.Int32, g.ID)
					if c, err := queries.GetWritingCategoryById(r.Context(), g.ItemID.Int32); err == nil && c.Title.Valid {
						gi.Info = c.Title.String
					}
				case "article":
					if w, err := queries.GetWritingForListerByID(r.Context(), db.GetWritingForListerByIDParams{ListerID: cd.UserID, Idwriting: g.ItemID.Int32, ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0}}); err == nil {
						if w.Title.Valid {
							info := w.Title.String
							if name, ok := langMap[w.LanguageIdlanguage]; ok && name != "" {
								info = fmt.Sprintf("[%s] %s", name, info)
							}
							gi.Info = info
						}
					}
				}
			case "faq":
				switch g.Item.String {
				case "category":
					if cats, err := queries.GetAllFAQCategories(r.Context()); err == nil {
						for _, c := range cats {
							if c.Idfaqcategories == g.ItemID.Int32 {
								if c.Name.Valid {
									gi.Info = c.Name.String
								}
								break
							}
						}
					}
				case "question", "question/answer":
					if qrow, err := queries.GetFAQByID(r.Context(), g.ItemID.Int32); err == nil && qrow.Question.Valid {
						text := qrow.Question.String
						if len(text) > 40 {
							text = text[:40] + "..."
						}
						if name, ok := langMap[qrow.LanguageIdlanguage]; ok && name != "" {
							text = fmt.Sprintf("[%s] %s", name, text)
						}
						gi.Info = text
					}
				}
			case "imagebbs":
				if g.Item.String == "board" {
					if b, err := queries.GetImageBoardById(r.Context(), g.ItemID.Int32); err == nil && b.Title.Valid {
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

	type GrantGroup struct {
		Section   string
		Item      string
		ItemID    sql.NullInt32
		Link      string
		Info      string
		Have      []string
		Available []string
	}

	actionMap := map[string][]string{
		"forum|topic":       {"see", "view", "reply", "post", "edit"},
		"forum|thread":      {"see", "view", "reply", "post", "edit"},
		"forum|category":    {"see", "view"},
		"linker|category":   {"see", "view"},
		"linker|link":       {"see", "view"},
		"images|upload":     {"see", "view", "post"},
		"news|post":         {"see", "view", "reply", "post", "edit"},
		"blog|category":     {"see", "view"},
		"blog|blog":         {"see", "view", "post", "edit"},
		"writings|category": {"see", "view"},
		"writings|writing":  {"see", "view", "post", "edit"},
	}

	groupMap := map[string]*GrantGroup{}
	for _, gi := range ginfos {
		key := fmt.Sprintf("%s|%s|%d", gi.Section, gi.Item.String, gi.ItemID.Int32)
		grp, ok := groupMap[key]
		if !ok {
			grp = &GrantGroup{Section: gi.Section, Item: gi.Item.String, ItemID: gi.ItemID, Link: gi.Link, Info: gi.Info}
			groupMap[key] = grp
		}
		grp.Have = append(grp.Have, gi.Action)
	}

	groups := make([]GrantGroup, 0, len(groupMap))
	for _, grp := range groupMap {
		if acts, ok := actionMap[grp.Section+"|"+grp.Item]; ok {
			haveSet := map[string]struct{}{}
			for _, h := range grp.Have {
				haveSet[h] = struct{}{}
			}
			for _, a := range acts {
				if _, ok := haveSet[a]; !ok {
					grp.Available = append(grp.Available, a)
				}
			}
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

	data := struct {
		*common.CoreData
		Role        *db.Role
		Users       []*db.AdminListUsersByRoleIDRow
		GrantGroups []GrantGroup
	}{
		CoreData:    cd,
		Role:        role,
		Users:       users,
		GrantGroups: groups,
	}

	handlers.TemplateHandler(w, r, "adminRolePage.gohtml", data)
}
