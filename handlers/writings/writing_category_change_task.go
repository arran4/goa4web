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

	if loop, err := writingCategoryWouldLoop(r.Context(), queries, int32(cid), int32(parentID)); err != nil {
		return fmt.Errorf("check writing category loop %w", handlers.ErrRedirectOnSamePageHandler(err))
	} else if loop {
		return common.UserError{ErrorMessage: "invalid parent category"}
	}

	if err := queries.UpdateWritingCategory(r.Context(), db.UpdateWritingCategoryParams{
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

func writingCategoryWouldLoop(ctx context.Context, queries *db.Queries, cid, parentID int32) (bool, error) {
	if parentID == 0 {
		return false, nil
	}
	if parentID == cid {
		return true, nil
	}
	cats, err := queries.FetchAllCategories(ctx)
	if err != nil {
		return false, err
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		parents[c.Idwritingcategory] = c.WritingCategoryID
	}
	seen := map[int32]struct{}{}
	for p := parentID; p != 0; {
		if _, ok := seen[p]; ok {
			return true, nil
		}
		seen[p] = struct{}{}
		if p == cid {
			return true, nil
		}
		np, ok := parents[p]
		if !ok {
			break
		}
		p = np
	}
	return false, nil
}
