package admin

import (
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/tasks"
)

var ResendQueueTask = resendQueueTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskResend,
		Match:     tasks.HasTask(TaskResend),
	},
}

var DeleteQueueTask = deleteQueueTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskDelete,
		Match:     tasks.HasTask(TaskDelete),
	},
}

var SaveTemplateTask = saveTemplateTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskUpdate,
		Match:     tasks.HasTask(TaskUpdate),
	},
}

var TestTemplateTask = testTemplateTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskTestMail,
		Match:     tasks.HasTask(TaskTestMail),
	},
}

var DeleteDLQTask = deleteDLQTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskDelete,
		Match:     tasks.HasTask(TaskDelete),
	},
}

var MarkReadTask = markReadTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskDismiss,
		Match:     tasks.HasTask(TaskDismiss),
	},
}

var PurgeNotificationsTask = purgeNotificationsTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskPurge,
		Match:     tasks.HasTask(TaskPurge),
	},
}

var SendNotificationTask = sendNotificationTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskNotify,
		Match:     tasks.HasTask(TaskNotify),
	},
}

var AddAnnouncementTask = addAnnouncementTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskAdd,
		Match:     tasks.HasTask(TaskAdd),
	},
}

var DeleteAnnouncementTask = deleteAnnouncementTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskDelete,
		Match:     tasks.HasTask(TaskDelete),
	},
}

var AddIPBanTask = addIPBanTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskAdd,
		Match:     tasks.HasTask(TaskAdd),
	},
}

var DeleteIPBanTask = deleteIPBanTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskDelete,
		Match:     tasks.HasTask(TaskDelete),
	},
}

var NewsUserAllowTask = tasks.BasicTaskEvent{
	EventName:     TaskAllow,
	Match:         tasks.HasTask(TaskAllow),
	ActionHandler: news.NewsAdminUserLevelsAllowActionPage,
}

var NewsUserRemoveTask = tasks.BasicTaskEvent{
	EventName:     TaskRemoveLower,
	Match:         tasks.HasTask(TaskRemoveLower),
	ActionHandler: news.NewsAdminUserLevelsRemoveActionPage,
}
