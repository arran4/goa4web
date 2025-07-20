package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
)

// ResendQueueTask triggers sending queued emails immediately.
type ResendQueueTask struct{ tasks.TaskString }

var resendQueueTask = &ResendQueueTask{TaskString: TaskResend}

var deleteQueueTask = &DeleteQueueTask{TaskString: TaskDelete}

var saveTemplateTask = &SaveTemplateTask{TaskString: TaskUpdate}

var testTemplateTask = &TestTemplateTask{TaskString: TaskTestMail}

var deleteDLQTask = &DeleteDLQTask{TaskString: TaskDelete}

var markReadTask = &MarkReadTask{TaskString: TaskDismiss}

var purgeNotificationsTask = &PurgeNotificationsTask{TaskString: TaskPurge}

var sendNotificationTask = &SendNotificationTask{TaskString: TaskNotify}

var addAnnouncementTask = &AddAnnouncementTask{TaskString: TaskAdd}

var deleteAnnouncementTask = &DeleteAnnouncementTask{TaskString: TaskDelete}

var addIPBanTask = &AddIPBanTask{TaskString: TaskAdd}

var deleteIPBanTask = &DeleteIPBanTask{TaskString: TaskDelete}

var NewsUserAllowTask = newsUserAllowTask{TaskString: tasks.TaskString("allow")}

var NewsUserRemoveTask = newsUserRemoveTask{TaskString: tasks.TaskString("remove")}
