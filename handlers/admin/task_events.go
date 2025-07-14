package admin

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/eventbus"
)

var ResendQueueTask = resendQueueTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskResend,
		Match:     hcommon.TaskMatcher(hcommon.TaskResend),
	},
}

var DeleteQueueTask = deleteQueueTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskDelete,
		Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	},
}

var SaveTemplateTask = saveTemplateTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskUpdate,
		Match:     hcommon.TaskMatcher(hcommon.TaskUpdate),
	},
}

var TestTemplateTask = testTemplateTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskTestMail,
		Match:     hcommon.TaskMatcher(hcommon.TaskTestMail),
	},
}

var DeleteDLQTask = deleteDLQTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskDelete,
		Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	},
}

var MarkReadTask = markReadTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskDismiss,
		Match:     hcommon.TaskMatcher(hcommon.TaskDismiss),
	},
}

var PurgeNotificationsTask = purgeNotificationsTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskPurge,
		Match:     hcommon.TaskMatcher(hcommon.TaskPurge),
	},
}

var SendNotificationTask = sendNotificationTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskNotify,
		Match:     hcommon.TaskMatcher(hcommon.TaskNotify),
	},
}

var AddAnnouncementTask = addAnnouncementTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskAdd,
		Match:     hcommon.TaskMatcher(hcommon.TaskAdd),
	},
}

var DeleteAnnouncementTask = deleteAnnouncementTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskDelete,
		Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	},
}

var AddIPBanTask = addIPBanTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskAdd,
		Match:     hcommon.TaskMatcher(hcommon.TaskAdd),
	},
}

var DeleteIPBanTask = deleteIPBanTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskDelete,
		Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	},
}

var NewsUserAllowTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskAllow,
	Match:         hcommon.TaskMatcher(hcommon.TaskAllow),
	ActionHandler: news.NewsAdminUserLevelsAllowActionPage,
}

var NewsUserRemoveTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskRemoveLower,
	Match:         hcommon.TaskMatcher(hcommon.TaskRemoveLower),
	ActionHandler: news.NewsAdminUserLevelsRemoveActionPage,
}
