package blogs

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks registers blog related tasks with the global registry.
func RegisterTasks() {
	tasks.Register(addBlogTask)
	tasks.Register(editBlogTask)
	tasks.Register(replyBlogTask)
	tasks.Register(editReplyTask)
	tasks.Register(cancelTask)
	tasks.Register(userAllowTask)
	tasks.Register(userDisallowTask)
	tasks.Register(usersAllowTask)
	tasks.Register(usersDisallowTask)
}
