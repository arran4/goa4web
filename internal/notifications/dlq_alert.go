package notifications

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// dlqAlertTask triggers admin notifications when the dead letter queue grows.
type dlqAlertTask struct{ tasks.TaskString }

func (dlqAlertTask) Action(http.ResponseWriter, *http.Request) {}

func (dlqAlertTask) AdminEmailTemplate() *EmailTemplates {
	return &EmailTemplates{
		Text: "dlqAlertEmail.gotxt",
		HTML: "dlqAlertEmail.gohtml",
	}
}

func (dlqAlertTask) AdminInternalNotificationTemplate() *string {
	s := "dlq_alert.gotxt"
	return &s
}
