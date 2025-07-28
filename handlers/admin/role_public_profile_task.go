package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// RolePublicProfileTask toggles public profile access for a role.
type RolePublicProfileTask struct{ tasks.TaskString }

var rolePublicProfileTask = &RolePublicProfileTask{TaskString: TaskToggleRolePublicProfile}

var _ tasks.Task = (*RolePublicProfileTask)(nil)
var _ tasks.AuditableTask = (*RolePublicProfileTask)(nil)

func (RolePublicProfileTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.PostFormValue("id"))
	enable := r.PostFormValue("enable") != ""
	var ts sql.NullTime
	if enable {
		ts = sql.NullTime{Time: time.Now(), Valid: true}
	}
	if err := queries.UpdateRolePublicProfileAllowed(r.Context(), db.UpdateRolePublicProfileAllowedParams{PublicProfileAllowedAt: ts, ID: int32(id)}); err != nil {
		return fmt.Errorf("update role fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["RoleID"] = id
			evt.Data["Enabled"] = enable
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: "/admin/roles"}
}

func (RolePublicProfileTask) AuditRecord(data map[string]any) string {
	id, _ := data["RoleID"].(int)
	if en, _ := data["Enabled"].(bool); en {
		return fmt.Sprintf("enabled public profiles for role %d", id)
	}
	return fmt.Sprintf("disabled public profiles for role %d", id)
}
