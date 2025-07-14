package user

import hcommon "github.com/arran4/goa4web/handlers/common"

var SaveLanguagesTask = hcommon.NewTaskEvent(TaskSaveLanguages)
var SaveLanguageTask = hcommon.NewTaskEvent(TaskSaveLanguage)
var SaveAllTask = hcommon.NewTaskEvent(TaskSaveAll)
var AddEmailTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var DeleteEmailTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var TestMailTask = hcommon.NewTaskEvent(TaskTestMail)
var DismissTask = hcommon.NewTaskEvent(TaskDismiss)
var SubscribeBlogsTask = hcommon.NewTaskEvent(hcommon.TaskSubscribeBlogs)
var SubscribeWritingsTask = hcommon.NewTaskEvent(hcommon.TaskSubscribeWritings)
var SubscribeNewsTask = hcommon.NewTaskEvent(hcommon.TaskSubscribeNews)
var SubscribeImagesTask = hcommon.NewTaskEvent(hcommon.TaskSubscribeImages)

// Permission management tasks used in the admin interface.
var PermissionUserAllowTask = hcommon.NewTaskEvent(hcommon.TaskUserAllow)
var PermissionUserDisallowTask = hcommon.NewTaskEvent(hcommon.TaskUserDisallow)
var PermissionUpdateTask = hcommon.NewTaskEvent(hcommon.TaskUpdate)

// DeleteTask removes a record such as a subscription.
var DeleteTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
