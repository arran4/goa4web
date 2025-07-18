package admin

import (
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/tasks"
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

// TODO move into this package even if it means updating link
var NewsUserAllowTask = news.NewsUserAllowTask{TaskString: TaskAllow}

// TODO move into this package even if it means updating link
var NewsUserRemoveTask = news.NewsUserRemoveTask{TaskString: TaskRemoveLower}
