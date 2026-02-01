package admin

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// DismissRequestTask dismisses a queued request.
type DismissRequestTask struct{ tasks.TaskString }

var dismissRequestTask = &DismissRequestTask{TaskString: TaskDismiss}

var _ tasks.Task = (*DismissRequestTask)(nil)
var _ tasks.AuditableTask = (*DismissRequestTask)(nil)

func (DismissRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "dismissed")
	return nil
}

// AuditRecord summarises a request queue action.
func (DismissRequestTask) AuditRecord(data map[string]any) string {
	return requestAuditSummary("dismissed", data)
}
