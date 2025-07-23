package admin

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// RejectRequestTask rejects a queued request.
type RejectRequestTask struct{ tasks.TaskString }

var rejectRequestTask = &RejectRequestTask{TaskString: TaskReject}

var _ tasks.Task = (*RejectRequestTask)(nil)
var _ tasks.AuditableTask = (*RejectRequestTask)(nil)

func (RejectRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "rejected")
	return nil
}

// AuditRecord summarises a request queue action.
func (RejectRequestTask) AuditRecord(data map[string]any) string {
	return requestAuditSummary("rejected", data)
}
