package user

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var SaveLanguagesTask = tasks.NewTaskEvent(TaskSaveLanguages)
var SaveLanguageTask = tasks.NewTaskEvent(TaskSaveLanguage)
var SaveAllTask = tasks.NewTaskEvent(TaskSaveAll)
var AddEmailTask = tasks.NewTaskEvent(TaskAdd)
var DeleteEmailTask = tasks.NewTaskEvent(TaskDelete)
var TestMailTask = tasks.NewTaskEvent(TaskTestMail)
var DismissTask = tasks.NewTaskEvent(TaskDismiss)
var UpdateSubscriptionsTask = tasks.NewTaskEvent(TaskUpdate)

// Permission management tasks used in the admin interface.
var PermissionUserAllowTask = tasks.NewTaskEvent(TaskUserAllow)
var PermissionUserDisallowTask = tasks.NewTaskEvent(TaskUserDisallow)
var PermissionUpdateTask = tasks.NewTaskEvent(TaskUpdate)

// DeleteTask removes a record such as a subscription.
var DeleteTask = tasks.NewTaskEvent(TaskDelete)
