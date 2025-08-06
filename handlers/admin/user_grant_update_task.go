package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserGrantUpdateTask updates the actions for a user on a specific item.
type UserGrantUpdateTask struct{ tasks.TaskString }

var userGrantUpdateTask = &UserGrantUpdateTask{TaskString: TaskRoleGrantUpdate}

var _ tasks.Task = (*UserGrantUpdateTask)(nil)

func (UserGrantUpdateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	userID := cd.CurrentProfileUserID()
	if userID == 0 {
		return fmt.Errorf("user id parse fail %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	section := r.PostFormValue("section")
	item := r.PostFormValue("item")
	itemIDStr := r.PostFormValue("item_id")
	actionsStr := r.PostFormValue("actions")
	disabledStr := r.PostFormValue("disabled_actions")
	if section == "" {
		return fmt.Errorf("missing section %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	var itemID sql.NullInt32
	if itemIDStr != "" {
		id, err := strconv.Atoi(itemIDStr)
		if err != nil {
			return fmt.Errorf("item id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		itemID = sql.NullInt32{Int32: int32(id), Valid: true}
	}
	if def, ok := GrantActionMap[section+"|"+item]; ok && def.RequireItemID && !itemID.Valid {
		return fmt.Errorf("missing item id %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	desiredActive := map[string]struct{}{}
	for _, a := range strings.Split(actionsStr, ",") {
		if a != "" {
			desiredActive[a] = struct{}{}
		}
	}
	desiredDisabled := map[string]struct{}{}
	for _, a := range strings.Split(disabledStr, ",") {
		if a != "" {
			desiredDisabled[a] = struct{}{}
		}
	}
	grants, err := queries.ListGrantsByUserID(r.Context(), sql.NullInt32{Int32: userID, Valid: true})
	if err != nil {
		return fmt.Errorf("list grants %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	existing := map[string]*db.Grant{}
	for _, g := range grants {
		if g.Section == section && g.Item.String == item && g.ItemID.Int32 == itemID.Int32 && g.ItemID.Valid == itemID.Valid {
			existing[g.Action] = g
		}
	}
	for a := range desiredActive {
		g, ok := existing[a]
		if !ok {
			if id, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Section:  section,
				Item:     sql.NullString{String: item, Valid: item != ""},
				RuleType: "allow",
				ItemID:   itemID,
				Action:   a,
			}); err != nil {
				return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			} else {
				g = &db.Grant{ID: int32(id)}
			}
		}
		if g != nil && !g.Active {
			if err := queries.AdminUpdateGrantActive(r.Context(), db.AdminUpdateGrantActiveParams{Active: true, ID: g.ID}); err != nil {
				return fmt.Errorf("activate grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}
	for a := range desiredDisabled {
		g, ok := existing[a]
		if !ok {
			id, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
				UserID:   sql.NullInt32{Int32: userID, Valid: true},
				Section:  section,
				Item:     sql.NullString{String: item, Valid: item != ""},
				RuleType: "allow",
				ItemID:   itemID,
				Action:   a,
			})
			if err != nil {
				return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			if err := queries.AdminUpdateGrantActive(r.Context(), db.AdminUpdateGrantActiveParams{Active: false, ID: int32(id)}); err != nil {
				return fmt.Errorf("deactivate grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		} else if g.Active {
			if err := queries.AdminUpdateGrantActive(r.Context(), db.AdminUpdateGrantActiveParams{Active: false, ID: g.ID}); err != nil {
				return fmt.Errorf("deactivate grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}
	for a, g := range existing {
		if _, ok := desiredActive[a]; ok {
			continue
		}
		if _, ok := desiredDisabled[a]; ok {
			continue
		}
		if err := queries.AdminDeleteGrant(r.Context(), g.ID); err != nil {
			return fmt.Errorf("delete grant %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d/grants", userID)}
}
