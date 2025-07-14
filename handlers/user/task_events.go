package user

import hcommon "github.com/arran4/goa4web/handlers/common"

var SaveLanguagesTask = hcommon.NewTaskEvent(TaskSaveLanguages)
var SaveLanguageTask = hcommon.NewTaskEvent(TaskSaveLanguage)
var SaveAllTask = hcommon.NewTaskEvent(TaskSaveAll)
var AddEmailTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var DeleteEmailTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var TestMailTask = hcommon.NewTaskEvent(TaskTestMail)
var DismissTask = hcommon.NewTaskEvent(TaskDismiss)
var UpdateSubscriptionsTask = hcommon.NewTaskEvent(hcommon.TaskUpdate)

// Permission management tasks used in the admin interface.
var PermissionUserAllowTask = hcommon.NewTaskEvent(hcommon.TaskUserAllow)
var PermissionUserDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUserDisallow)
var PermissionUpdateTask = hcommon.NewTaskEvent(hcommon.TaskUpdate)

// DeleteTask removes a record such as a subscription.
var DeleteTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
