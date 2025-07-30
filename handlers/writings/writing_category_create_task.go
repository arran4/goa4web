package writings

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// WritingCategoryCreateTask creates a new category.
type WritingCategoryCreateTask struct{ tasks.TaskString }

var writingCategoryCreateTask = &WritingCategoryCreateTask{TaskString: TaskWritingCategoryCreate}

var _ tasks.Task = (*WritingCategoryCreateTask)(nil)

func (WritingCategoryCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		return fmt.Errorf("pcid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cats, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		return fmt.Errorf("fetch categories %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idwritingcategory] = c.WritingCategoryID
	}
	if path, loop := algorithms.WouldCreateLoop(parents, 0, int32(pcid)); loop && len(path) > 0 {
		return common.UserError{ErrorMessage: "invalid parent category: loop detected"}
	}
	if err := queries.InsertWritingCategory(r.Context(), db.InsertWritingCategoryParams{
		WritingCategoryID: int32(pcid),
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
	}); err != nil {
		return fmt.Errorf("create writing category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
