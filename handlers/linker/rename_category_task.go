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

type renameCategoryTask struct{ tasks.TaskString }

var RenameCategoryTask = &renameCategoryTask{TaskString: TaskRenameCategory}
var _ tasks.Task = (*renameCategoryTask)(nil)

func (renameCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, _ := strconv.Atoi(r.PostFormValue("cid"))
	title := r.PostFormValue("title")
	pos, _ := strconv.Atoi(r.PostFormValue("position"))
	if err := queries.RenameLinkerCategory(r.Context(), db.RenameLinkerCategoryParams{
		Title:            sql.NullString{Valid: true, String: title},
		Position:         int32(pos),
		Idlinkercategory: int32(cid),
	}); err != nil {
		return fmt.Errorf("rename linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
