package linker

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type updateCategoryTask struct{ tasks.TaskString }

var UpdateCategoryTask = &updateCategoryTask{TaskString: TaskUpdate}
var _ tasks.Task = (*updateCategoryTask)(nil)

func (updateCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.AdminRenameLinkerCategory(r.Context(), db.AdminRenameLinkerCategoryParams{
		Title:    sql.NullString{Valid: true, String: title},
		Position: int32(pos),
		ID:       int32(cid),
	}); err != nil {
		return fmt.Errorf("rename linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	order, _ := strconv.Atoi(r.PostFormValue("order"))
	if err := queries.AdminUpdateLinkerCategorySortOrder(r.Context(), db.AdminUpdateLinkerCategorySortOrderParams{
		Sortorder: int32(order),
		ID:        int32(cid),
	}); err != nil {
		return fmt.Errorf("update linker category sort order fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
