package admin

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// AcceptRequestTask accepts a queued request.
type AcceptRequestTask struct{ tasks.TaskString }

var acceptRequestTask = &AcceptRequestTask{TaskString: TaskAccept}

var _ tasks.Task = (*AcceptRequestTask)(nil)
var _ tasks.AuditableTask = (*AcceptRequestTask)(nil)

func (AcceptRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "accepted")
	return nil
}

// AuditRecord summarises a request queue action.
func (AcceptRequestTask) AuditRecord(data map[string]any) string {
	return requestAuditSummary("accepted", data)
}
