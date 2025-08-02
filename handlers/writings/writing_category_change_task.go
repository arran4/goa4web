package writings

import (
	"context"
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

// WritingCategoryChangeTask modifies a category.
type WritingCategoryChangeTask struct{ tasks.TaskString }

var writingCategoryChangeTask = &WritingCategoryChangeTask{TaskString: TaskWritingCategoryChange}

var _ tasks.Task = (*WritingCategoryChangeTask)(nil)

func (WritingCategoryChangeTask) Action(w http.ResponseWriter, r *http.Request) any {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	parentID, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		return fmt.Errorf("pcid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if path, loop, err := writingCategoryWouldLoop(r.Context(), queries, int32(cid), int32(parentID)); err != nil {
		return fmt.Errorf("check writing category loop %w", handlers.ErrRedirectOnSamePageHandler(err))
	} else if loop {
		return common.UserError{ErrorMessage: fmt.Sprintf("invalid parent category: loop %v", path)}
	}

	if err := queries.AdminUpdateWritingCategory(r.Context(), db.AdminUpdateWritingCategoryParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idwritingcategory: int32(cid),
		WritingCategoryID: int32(parentID),
	}); err != nil {
		return fmt.Errorf("update writing category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return nil
}

func writingCategoryWouldLoop(ctx context.Context, queries *db.Queries, cid, parentID int32) ([]int32, bool, error) {
	var parents map[int32]int32
	if queries != nil {
		cats, err := queries.FetchAllCategories(ctx)
		if err != nil {
			return nil, false, err
		}
		parents = make(map[int32]int32, len(cats))
		for _, c := range cats {
			parents[c.Idwritingcategory] = c.WritingCategoryID
		}
	} else {
		parents = map[int32]int32{}
	}
	path, loop := algorithms.WouldCreateLoop(parents, cid, parentID)
	return path, loop, nil
}
