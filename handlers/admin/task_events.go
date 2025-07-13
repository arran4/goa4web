package admin

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/eventbus"
)

var ResendQueueTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskResend,
	Match:     hcommon.TaskMatcher(hcommon.TaskResend),
	ActionH:   AdminEmailQueueResendActionPage,
}

var DeleteQueueTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDelete,
	Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionH:   AdminEmailQueueDeleteActionPage,
}

var SaveTemplateTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskUpdate,
	Match:     hcommon.TaskMatcher(hcommon.TaskUpdate),
	ActionH:   AdminEmailTemplateSaveActionPage,
}

var TestTemplateTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskTestMail,
	Match:     hcommon.TaskMatcher(hcommon.TaskTestMail),
	ActionH:   AdminEmailTemplateTestActionPage,
}

var DeleteDLQTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDelete,
	Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionH:   AdminDLQAction,
}

var MarkReadTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDismiss,
	Match:     hcommon.TaskMatcher(hcommon.TaskDismiss),
	ActionH:   AdminNotificationsMarkReadActionPage,
}

var PurgeNotificationsTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskPurge,
	Match:     hcommon.TaskMatcher(hcommon.TaskPurge),
	ActionH:   AdminNotificationsPurgeActionPage,
}

var SendNotificationTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskNotify,
	Match:     hcommon.TaskMatcher(hcommon.TaskNotify),
	ActionH:   AdminNotificationsSendActionPage,
}

var AddAnnouncementTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskAdd,
	Match:     hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionH:   AdminAnnouncementsAddActionPage,
}

var DeleteAnnouncementTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDelete,
	Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionH:   AdminAnnouncementsDeleteActionPage,
}

var AddIPBanTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskAdd,
	Match:     hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionH:   AdminIPBanAddActionPage,
}

var DeleteIPBanTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskDelete,
	Match:     hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionH:   AdminIPBanDeleteActionPage,
}

var NewsUserAllowTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskAllow,
	Match:     hcommon.TaskMatcher(hcommon.TaskAllow),
	ActionH:   news.NewsAdminUserLevelsAllowActionPage,
}

var NewsUserRemoveTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskRemoveLower,
	Match:     hcommon.TaskMatcher(hcommon.TaskRemoveLower),
	ActionH:   news.NewsAdminUserLevelsRemoveActionPage,
}
