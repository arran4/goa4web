package admin

import (
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/tasks"
)

var ResendQueueTask = resendQueueTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskResend,
		Match:     tasks.HasTask(tasks.TaskResend),
	},
}

var DeleteQueueTask = deleteQueueTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskDelete,
		Match:     tasks.HasTask(tasks.TaskDelete),
	},
}

var SaveTemplateTask = saveTemplateTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskUpdate,
		Match:     tasks.HasTask(tasks.TaskUpdate),
	},
}

var TestTemplateTask = testTemplateTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskTestMail,
		Match:     tasks.HasTask(tasks.TaskTestMail),
	},
}

var DeleteDLQTask = deleteDLQTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskDelete,
		Match:     tasks.HasTask(tasks.TaskDelete),
	},
}

var MarkReadTask = markReadTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskDismiss,
		Match:     tasks.HasTask(tasks.TaskDismiss),
	},
}

var PurgeNotificationsTask = purgeNotificationsTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskPurge,
		Match:     tasks.HasTask(tasks.TaskPurge),
	},
}

var SendNotificationTask = sendNotificationTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskNotify,
		Match:     tasks.HasTask(tasks.TaskNotify),
	},
}

var AddAnnouncementTask = addAnnouncementTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskAdd,
		Match:     tasks.HasTask(tasks.TaskAdd),
	},
}

var DeleteAnnouncementTask = deleteAnnouncementTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskDelete,
		Match:     tasks.HasTask(tasks.TaskDelete),
	},
}

var AddIPBanTask = addIPBanTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskAdd,
		Match:     tasks.HasTask(tasks.TaskAdd),
	},
}

var DeleteIPBanTask = deleteIPBanTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: tasks.TaskDelete,
		Match:     tasks.HasTask(tasks.TaskDelete),
	},
}

var NewsUserAllowTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskAllow,
	Match:         tasks.HasTask(tasks.TaskAllow),
	ActionHandler: news.NewsAdminUserLevelsAllowActionPage,
}

var NewsUserRemoveTask = tasks.BasicTaskEvent{
	EventName:     tasks.TaskRemoveLower,
	Match:         tasks.HasTask(tasks.TaskRemoveLower),
	ActionHandler: news.NewsAdminUserLevelsRemoveActionPage,
}
