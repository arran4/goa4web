package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// CategoryGrantCreateTask creates a new grant for a forum category.
type CategoryGrantCreateTask struct{ tasks.TaskString }

var categoryGrantCreateTask = &CategoryGrantCreateTask{TaskString: forumcommon.TaskCategoryGrantCreate}

var _ tasks.Task = (*CategoryGrantCreateTask)(nil)

func (CategoryGrantCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["category"])
	if err != nil {
		return fmt.Errorf("category id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	username := r.PostFormValue("username")
	role := r.PostFormValue("role")
	actions := r.Form["action"]
	if len(actions) == 0 {
		actions = []string{"see"}
	}
	var uid sql.NullInt32
	if username != "" {
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("SystemGetUserByUsername: %v", err)
			return fmt.Errorf("get user by username %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		uid = sql.NullInt32{Int32: u.Idusers, Valid: true}
	}
	var rid sql.NullInt32
	if role != "" {
		roles, err := queries.AdminListRoles(r.Context())
		if err != nil {
			log.Printf("ListRoles: %v", err)
			return fmt.Errorf("list roles %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		for _, ro := range roles {
			if ro.Name == role {
				rid = sql.NullInt32{Int32: ro.ID, Valid: true}
				break
			}
		}
	}
	for _, action := range actions {
		if action == "" {
			action = "see"
		}
		if _, err = cd.GrantForumCategory(int32(categoryID), uid, rid, action); err != nil {
			log.Printf("CreateGrant: %v", err)
			return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/forum/categories/category/%d/grants", categoryID)}
}
