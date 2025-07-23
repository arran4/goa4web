package admin

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// QueryRequestTask asks for more information about a request.
type QueryRequestTask struct{ tasks.TaskString }

var queryRequestTask = &QueryRequestTask{TaskString: TaskQuery}

var _ tasks.Task = (*QueryRequestTask)(nil)
var _ tasks.AuditableTask = (*QueryRequestTask)(nil)

func (QueryRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	handleRequestAction(w, r, "query")
	return nil
}

// AuditRecord summarises a request queue action.
func (QueryRequestTask) AuditRecord(data map[string]any) string {
	return requestAuditSummary("query", data)
}
