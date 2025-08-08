package writings

import (
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

// CategoryGrantCreateTask creates a new grant for a writing category.
type CategoryGrantCreateTask struct{ tasks.TaskString }

var categoryGrantCreateTask = &CategoryGrantCreateTask{TaskString: TaskCategoryGrantCreate}

var _ tasks.Task = (*CategoryGrantCreateTask)(nil)

func (CategoryGrantCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
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
	var uid, rid int32
	if username != "" {
		if u, err := cd.WriterByUsername(username); err == nil {
			uid = u.Idusers
		} else {
			log.Printf("SystemGetUserByUsername: %v", err)
			return fmt.Errorf("get user by username %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	if role != "" {
		roles, err := cd.Queries().AdminListRoles(r.Context())
		if err != nil {
			log.Printf("ListRoles: %v", err)
			return fmt.Errorf("list roles %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		for _, ro := range roles {
			if ro.Name == role {
				rid = ro.ID
				break
			}
		}
	}
	if err := cd.GrantWritingCategory(int32(categoryID), uid, rid, actions); err != nil {
		log.Printf("CreateGrant: %v", err)
		return fmt.Errorf("create grant %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
