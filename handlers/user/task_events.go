package user

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var SaveLanguagesTask = tasks.NewTaskEvent(TaskSaveLanguages)
var SaveLanguageTask = tasks.NewTaskEvent(TaskSaveLanguage)
var SaveAllTask = tasks.NewTaskEvent(TaskSaveAll)
var AddEmailTask = tasks.NewTaskEvent(tasks.TaskAdd)
var DeleteEmailTask = tasks.NewTaskEvent(tasks.TaskDelete)
var TestMailTask = tasks.NewTaskEvent(TaskTestMail)
var DismissTask = tasks.NewTaskEvent(TaskDismiss)
var UpdateSubscriptionsTask = tasks.NewTaskEvent(tasks.TaskUpdate)

// Permission management tasks used in the admin interface.
var PermissionUserAllowTask = tasks.NewTaskEvent(tasks.TaskUserAllow)
var PermissionUserDisallowTask = tasks.NewTaskEvent(tasks.TaskUserDisallow)
var PermissionUpdateTask = tasks.NewTaskEvent(tasks.TaskUpdate)

// DeleteTask removes a record such as a subscription.
var DeleteTask = tasks.NewTaskEvent(tasks.TaskDelete)
