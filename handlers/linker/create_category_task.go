package linker

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type createCategoryTask struct{ tasks.TaskString }

var CreateCategoryTask = &createCategoryTask{TaskString: TaskCreateCategory}
var _ tasks.Task = (*createCategoryTask)(nil)

func (createCategoryTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	title := r.PostFormValue("title")
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	rows, _ := cd.LinkerCategoryCounts()
	pos := len(rows) + 1
	if err := queries.AdminCreateLinkerCategory(r.Context(), db.AdminCreateLinkerCategoryParams{
		Title:    sql.NullString{Valid: true, String: title},
		Position: int32(pos),
	}); err != nil {
		return fmt.Errorf("create linker category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
