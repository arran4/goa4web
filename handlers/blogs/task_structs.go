package blogs

import (
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"
)

// EditReplyTask updates an existing comment.
type EditReplyTask struct{ tasks.TaskString }

var editReplyTask = &EditReplyTask{TaskString: TaskEditReply}

func (EditReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	CommentEditPostPage(w, r)
}

// CancelTask cancels comment editing.
type CancelTask struct{ tasks.TaskString }

var cancelTask = &CancelTask{TaskString: TaskCancel}

func (CancelTask) Action(w http.ResponseWriter, r *http.Request) {
	CommentEditPostCancelPage(w, r)
}
