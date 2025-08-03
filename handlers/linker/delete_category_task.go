package linker

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type deleteCategoryTask struct{ tasks.TaskString }

var AdminDeleteCategoryTask = &deleteCategoryTask{TaskString: TaskDeleteCategory}
var _ tasks.Task = (*deleteCategoryTask)(nil)

func (deleteCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	rows, _ := cd.LinkerCategoryCounts()
	for _, c := range rows {
		if int(c.Idlinkercategory) == cid && c.Linkcount > 0 {
			http.Error(w, "Category in use", http.StatusBadRequest)
			return nil
		}
	}
	count, err := queries.AdminCountLinksByCategory(r.Context(), int32(cid))
	if err != nil {
		return fmt.Errorf("count links fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if count > 0 {
		http.Error(w, "Category in use", http.StatusBadRequest)
		return nil
	}
	if err := queries.AdminDeleteLinkerCategory(r.Context(), int32(cid)); err != nil {
		return fmt.Errorf("delete linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
