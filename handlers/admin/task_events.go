package admin

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/eventbus"
)

var ResendQueueTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskResend,
	Match:         hcommon.TaskMatcher(hcommon.TaskResend),
	ActionHandler: AdminEmailQueueResendActionPage,
}

var DeleteQueueTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionHandler: AdminEmailQueueDeleteActionPage,
}

var SaveTemplateTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskUpdate,
	Match:         hcommon.TaskMatcher(hcommon.TaskUpdate),
	ActionHandler: AdminEmailTemplateSaveActionPage,
}

var TestTemplateTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskTestMail,
	Match:         hcommon.TaskMatcher(hcommon.TaskTestMail),
	ActionHandler: AdminEmailTemplateTestActionPage,
}

var DeleteDLQTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionHandler: AdminDLQAction,
}

var MarkReadTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDismiss,
	Match:         hcommon.TaskMatcher(hcommon.TaskDismiss),
	ActionHandler: AdminNotificationsMarkReadActionPage,
}

var PurgeNotificationsTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskPurge,
	Match:         hcommon.TaskMatcher(hcommon.TaskPurge),
	ActionHandler: AdminNotificationsPurgeActionPage,
}

var SendNotificationTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskNotify,
	Match:         hcommon.TaskMatcher(hcommon.TaskNotify),
	ActionHandler: AdminNotificationsSendActionPage,
}

var AddAnnouncementTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskAdd,
	Match:         hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionHandler: AdminAnnouncementsAddActionPage,
}

var DeleteAnnouncementTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionHandler: AdminAnnouncementsDeleteActionPage,
}

var AddIPBanTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskAdd,
	Match:         hcommon.TaskMatcher(hcommon.TaskAdd),
	ActionHandler: AdminIPBanAddActionPage,
}

var DeleteIPBanTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskDelete,
	Match:         hcommon.TaskMatcher(hcommon.TaskDelete),
	ActionHandler: AdminIPBanDeleteActionPage,
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
