package admin

import (
	"github.com/arran4/goa4web/handlers/news"
)

var ResendQueueTask = resendQueueTask{TaskString: TaskResend}

var DeleteQueueTask = deleteQueueTask{TaskString: TaskDelete}

var SaveTemplateTask = saveTemplateTask{TaskString: TaskUpdate}

var TestTemplateTask = testTemplateTask{TaskString: TaskTestMail}

var DeleteDLQTask = deleteDLQTask{TaskString: TaskDelete}

var MarkReadTask = markReadTask{TaskString: TaskDismiss}

var PurgeNotificationsTask = purgeNotificationsTask{TaskString: TaskPurge}

var SendNotificationTask = sendNotificationTask{TaskString: TaskNotify}

var AddAnnouncementTask = addAnnouncementTask{TaskString: TaskAdd}

var DeleteAnnouncementTask = deleteAnnouncementTask{TaskString: TaskDelete}

var AddIPBanTask = addIPBanTask{TaskString: TaskAdd}

var DeleteIPBanTask = deleteIPBanTask{TaskString: TaskDelete}

var NewsUserAllowTask = news.NewsUserAllowTask

var NewsUserRemoveTask = news.NewsUserRemoveTask
