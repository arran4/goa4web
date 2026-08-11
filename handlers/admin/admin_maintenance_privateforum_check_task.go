package admin

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type CheckPrivateForumGrantsTask struct{ tasks.TaskString }

var checkPrivateForumGrantsTask = &CheckPrivateForumGrantsTask{TaskString: TaskCheckPrivateForumGrants}

var _ tasks.Task = (*CheckPrivateForumGrantsTask)(nil)

type checkPrivateForumGrantsData struct {
	Inconsistencies []common.PrivateForumInconsistency
	IsPreview       bool
	TaskName        string
}

func (CheckPrivateForumGrantsTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	dryRun := r.FormValue("preview") == "true"

	// Collect fix IDs
	var fixIDs []string
	if !dryRun {
		fixIDs = r.Form["fix_id"]
	}

	inconsistencies, err := cd.CheckAndFixPrivateForumInconsistencies(r.Context(), fixIDs, dryRun)
	if err != nil {
		return fmt.Errorf("checking private forum inconsistencies: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	_ = AdminMaintenanceCheckPrivateForumPreviewPageTmpl.Handle(w, r, checkPrivateForumGrantsData{
		Inconsistencies: inconsistencies,
		IsPreview:       dryRun,
		TaskName:        string(TaskCheckPrivateForumGrants),
	})
	return nil
}

const AdminMaintenanceCheckPrivateForumPreviewPageTmpl tasks.Template = "domains/admin/maintenancePrivateForumCheckPreviewPage.gohtml"
