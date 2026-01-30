package dlq

import (
	"github.com/arran4/goa4web/internal/eventbus"
)

// Message represents a message stored in the DLQ.
type Message struct {
	Error    string              `json:"error"`
	Event    *eventbus.TaskEvent `json:"event,omitempty"`
	TaskName string              `json:"task_name,omitempty"`
}
