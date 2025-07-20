package admin

import "github.com/arran4/goa4web/internal/tasks"

// TODO var resendQueueTask = ResendQueueTask{
var ResendQueueTask = resendQueueTask{TaskString: TaskResend}

// TODO var deleteQueueTask = DeleteQueueTask{
var DeleteQueueTask = deleteQueueTask{TaskString: TaskDelete}

// TODO var saveTemplateTask = SaveTemplateTask{
var SaveTemplateTask = saveTemplateTask{TaskString: TaskUpdate}

// TODO var testTemplateTask = TestTemplateTask{
var TestTemplateTask = testTemplateTask{TaskString: TaskTestMail}

// TODO var deleteDLQTask = DeleteDLQTask{
var DeleteDLQTask = deleteDLQTask{TaskString: TaskDelete}

// TODO var markReadTask = MarkReadTask{
var MarkReadTask = markReadTask{TaskString: TaskDismiss}

// TODO var purgeNotificationsTask = PurgeNotificationsTask{
var PurgeNotificationsTask = purgeNotificationsTask{TaskString: TaskPurge}

// TODO var sendNotificationTask = SendNotificationTask{
var SendNotificationTask = sendNotificationTask{TaskString: TaskNotify}

// TODO var addAnnouncementTask = AddAnnouncementTask{
var AddAnnouncementTask = addAnnouncementTask{TaskString: TaskAdd}

// TODO var deleteAnnouncementTask = DeleteAnnouncementTask{
var DeleteAnnouncementTask = deleteAnnouncementTask{TaskString: TaskDelete}

// TODO var addIPBanTask = AddIPBanTask{
var AddIPBanTask = addIPBanTask{TaskString: TaskAdd}

var DeleteIPBanTask = deleteIPBanTask{TaskString: TaskDelete}

var NewsUserAllowTask = newsUserAllowTask{TaskString: tasks.TaskString("allow")}

var NewsUserRemoveTask = newsUserRemoveTask{TaskString: tasks.TaskString("remove")}
