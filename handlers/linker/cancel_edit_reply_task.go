package linker

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

type cancelEditReplyTask struct{ tasks.TaskString }

var commentEditActionCancel = &cancelEditReplyTask{TaskString: TaskCancel}
var _ tasks.Task = (*cancelEditReplyTask)(nil)

func (cancelEditReplyTask) Page(w http.ResponseWriter, r *http.Request) {
	CommentEditActionCancelPage(w, r)
}
