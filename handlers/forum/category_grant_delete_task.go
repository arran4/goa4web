package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
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

// CategoryGrantDeleteTask removes a grant from a forum category.
type CategoryGrantDeleteTask struct{ tasks.TaskString }

var categoryGrantDeleteTask = &CategoryGrantDeleteTask{TaskString: forumcommon.TaskCategoryGrantDelete}

var _ tasks.Task = (*CategoryGrantDeleteTask)(nil)

func (CategoryGrantDeleteTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["category"])
	if err != nil {
		return fmt.Errorf("category id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	grantID, err := strconv.Atoi(r.PostFormValue("grantid"))
	if err != nil {
		return fmt.Errorf("grant id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.RevokeForumCategory(int32(grantID)); err != nil {
		log.Printf("DeleteGrant: %v", err)
		return fmt.Errorf("delete grant %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/forum/categories/category/%d/grants", categoryID)}
}
