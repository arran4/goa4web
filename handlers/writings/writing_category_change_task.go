package writings

import (
	"context"
	"fmt"
	"math"
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		return fmt.Errorf("cid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.ChangeWritingCategory(int32(cid), int32(parentID), name, desc); err != nil {
		if _, ok := err.(common.UserError); ok {
			return err
		}
		return fmt.Errorf("update writing category fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

func writingCategoryWouldLoop(ctx context.Context, queries db.Querier, cid, parentID int32) ([]int32, bool, error) {
	var parents map[int32]int32
	if queries != nil {
		cats, err := queries.SystemListWritingCategories(ctx, db.SystemListWritingCategoriesParams{Limit: math.MaxInt32, Offset: 0})
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
