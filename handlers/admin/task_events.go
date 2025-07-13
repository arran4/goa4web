package admin

import hcommon "github.com/arran4/goa4web/handlers/common"

var ResendQueueTask = hcommon.NewTaskEvent(hcommon.TaskResend)
var DeleteQueueTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var SaveTemplateTask = hcommon.NewTaskEvent(hcommon.TaskUpdate)
var TestTemplateTask = hcommon.NewTaskEvent(hcommon.TaskTestMail)
var DeleteDLQTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var MarkReadTask = hcommon.NewTaskEvent(hcommon.TaskDismiss)
var PurgeNotificationsTask = hcommon.NewTaskEvent(hcommon.TaskPurge)
var SendNotificationTask = hcommon.NewTaskEvent(hcommon.TaskNotify)
var AddAnnouncementTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var DeleteAnnouncementTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var AddIPBanTask = hcommon.NewTaskEvent(hcommon.TaskAdd)
var DeleteIPBanTask = hcommon.NewTaskEvent(hcommon.TaskDelete)
var NewsUserAllowTask = hcommon.NewTaskEvent(hcommon.TaskAllow)
var NewsUserRemoveTask = hcommon.NewTaskEvent(hcommon.TaskRemoveLower)
