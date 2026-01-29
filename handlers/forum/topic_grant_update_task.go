package forum

import (
	"github.com/arran4/goa4web/handlers/forum/forumcommon"
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
	"github.com/gorilla/mux"
)

// TopicGrantUpdateTask updates role grants for a forum topic action.
type TopicGrantUpdateTask struct{ tasks.TaskString }

var topicGrantUpdateTask = &TopicGrantUpdateTask{TaskString: forumcommon.TaskTopicGrantUpdate}

var _ tasks.Task = (*TopicGrantUpdateTask)(nil)

func (TopicGrantUpdateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	tid, err := strconv.Atoi(mux.Vars(r)["topic"])
	if err != nil {
		return fmt.Errorf("topic id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	action := r.PostFormValue("action")
	activeStr := r.PostFormValue("actions")
	disabledStr := r.PostFormValue("disabled_actions")
	desiredActive := map[string]struct{}{}
	for _, n := range strings.Split(activeStr, ",") {
		if n != "" {
			desiredActive[n] = struct{}{}
		}
	}
	desiredDisabled := map[string]struct{}{}
	for _, n := range strings.Split(disabledStr, ",") {
		if n != "" {
			desiredDisabled[n] = struct{}{}
		}
	}
	roles, err := cd.AllRoles()
	if err != nil {
		return fmt.Errorf("list roles %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	roleIDs := map[string]int32{}
	idToName := map[int32]string{}
	for _, ro := range roles {
		roleIDs[ro.Name] = ro.ID
		idToName[ro.ID] = ro.Name
	}
	grants, err := queries.ListGrants(r.Context())
	if err != nil {
		return fmt.Errorf("list grants %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	existing := map[int32]*db.Grant{}
	for _, g := range grants {
		if g.Section == "forum" && g.Item.Valid && g.Item.String == "topic" && g.ItemID.Valid && g.ItemID.Int32 == int32(tid) && g.Action == action && !g.UserID.Valid {
			rid := int32(0)
			if g.RoleID.Valid {
				rid = g.RoleID.Int32
			}
			existing[rid] = g
		}
	}
	for name := range desiredActive {
		rid, ok := roleIDs[name]
		if !ok {
			continue
		}
		g, ok := existing[rid]
		if !ok {
			id, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
				RoleID:   sql.NullInt32{Int32: rid, Valid: rid != 0},
				Section:  "forum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: int32(tid), Valid: true},
				Action:   action,
			})
			if err != nil {
				return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			g = &db.Grant{ID: int32(id)}
		}
		if g != nil && !g.Active {
			if err := queries.AdminUpdateGrantActive(r.Context(), db.AdminUpdateGrantActiveParams{Active: true, ID: g.ID}); err != nil {
				return fmt.Errorf("activate grant %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
	}
	for name := range desiredDisabled {
		rid, ok := roleIDs[name]
		if !ok {
			continue
		}
		g, ok := existing[rid]
		if !ok {
			id, err := queries.AdminCreateGrant(r.Context(), db.AdminCreateGrantParams{
				RoleID:   sql.NullInt32{Int32: rid, Valid: rid != 0},
				Section:  "forum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: int32(tid), Valid: true},
				Action:   action,
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
	for id, g := range existing {
		name := idToName[id]
		if _, ok := desiredActive[name]; ok {
			continue
		}
		if _, ok := desiredDisabled[name]; ok {
			continue
		}
		if err := queries.AdminDeleteGrant(r.Context(), g.ID); err != nil {
			return fmt.Errorf("delete grant %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/forum/topics/topic/%d/grants", tid)}
}
