package user

import (
	"github.com/arran4/goa4web/internal/tasks"
)

var SaveLanguagesEvent = tasks.NewTaskEvent(TaskSaveLanguages)
var SaveLanguageEvent = tasks.NewTaskEvent(TaskSaveLanguage)
var SaveAllEvent = tasks.NewTaskEvent(TaskSaveAll)
var AddEmailEvent = tasks.NewTaskEvent(TaskAdd)
var DeleteEmailEvent = tasks.NewTaskEvent(TaskDelete)
var TestMailEvent = tasks.NewTaskEvent(TaskTestMail)
var DismissEvent = tasks.NewTaskEvent(TaskDismiss)
var UpdateSubscriptionsEvent = tasks.NewTaskEvent(TaskUpdate)

// Permission management tasks used in the admin interface.
var PermissionUserAllowEvent = tasks.NewTaskEvent(TaskUserAllow)
var PermissionUserDisallowEvent = tasks.NewTaskEvent(TaskUserDisallow)
var PermissionUpdateEvent = tasks.NewTaskEvent(TaskUpdate)

// DeleteTask removes a record such as a subscription.
var DeleteEvent = tasks.NewTaskEvent(TaskDelete)
