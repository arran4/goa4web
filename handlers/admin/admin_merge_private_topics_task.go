package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type MergePrivateTopicsTask struct{ tasks.TaskString }

var mergePrivateTopicsTask = &MergePrivateTopicsTask{TaskString: TaskMergePrivateTopics}

var _ tasks.Task = (*MergePrivateTopicsTask)(nil)

func (MergePrivateTopicsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	_, err := cd.MergePrivateTopicsWithSameParticipants(r.Context(), false)
	if err != nil {
		return fmt.Errorf("merging private topics: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/maintenance"}
}
