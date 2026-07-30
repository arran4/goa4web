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

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	dryRun := r.FormValue("preview") == "true"

	mergedCount, err := cd.MergePrivateTopicsWithSameParticipants(r.Context(), dryRun)
	if err != nil {
		return fmt.Errorf("merging private topics: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if dryRun {
		data := struct{
			Message string
		}{
			Message: fmt.Sprintf("Preview completed. Would merge %d topics.", mergedCount),
		}
		AdminMaintenancePreviewPageTmpl.Handle(w, r, data)
		return nil
	}

	return handlers.RefreshDirectHandler{TargetURL: "/admin/maintenance"}
}

const AdminMaintenancePreviewPageTmpl tasks.Template = "admin/maintenancePreviewPage.gohtml"
