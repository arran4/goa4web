package linker

import "github.com/arran4/goa4web/internal/tasks"

type cancelEditReplyTask struct{ tasks.TaskString }

var commentEditActionCancel = &cancelEditReplyTask{TaskString: TaskCancel}
var _ tasks.Task = (*cancelEditReplyTask)(nil)
