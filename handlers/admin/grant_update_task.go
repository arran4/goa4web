package admin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// GrantUpdateTask updates the active flag on a grant.
type GrantUpdateTask struct{ tasks.TaskString }

var grantUpdateTask = &GrantUpdateTask{TaskString: TaskGrantUpdateActive}

var _ tasks.Task = (*GrantUpdateTask)(nil)

func (GrantUpdateTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, err := strconv.Atoi(r.PostFormValue("grantid"))
	if err != nil {
		return fmt.Errorf("grant id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	active := r.PostFormValue("active") == "1"
	if err := queries.AdminUpdateGrantActive(r.Context(), db.AdminUpdateGrantActiveParams{Active: active, ID: int32(id)}); err != nil {
		log.Printf("UpdateGrant: %v", err)
		return fmt.Errorf("update grant %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/grant/%d", id)}
}
