package admin

import (
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/eventbus"
)

// task types with receiver-based actions
type resendQueueTask struct{ eventbus.BasicTaskEvent }
type deleteQueueTask struct{ eventbus.BasicTaskEvent }
type saveTemplateTask struct{ eventbus.BasicTaskEvent }
type testTemplateTask struct{ eventbus.BasicTaskEvent }
type deleteDLQTask struct{ eventbus.BasicTaskEvent }
type markReadTask struct{ eventbus.BasicTaskEvent }
type purgeNotificationsTask struct{ eventbus.BasicTaskEvent }
type sendNotificationTask struct{ eventbus.BasicTaskEvent }
type addAnnouncementTask struct{ eventbus.BasicTaskEvent }
type deleteAnnouncementTask struct{ eventbus.BasicTaskEvent }
type addIPBanTask struct{ eventbus.BasicTaskEvent }
type deleteIPBanTask struct{ eventbus.BasicTaskEvent }
type newsUserAllowTask struct{ eventbus.BasicTaskEvent }
type newsUserRemoveTask struct{ eventbus.BasicTaskEvent }

func (resendQueueTask) Action() http.HandlerFunc        { return ResendQueueTask.action }
func (resendQueueTask) Page() http.HandlerFunc          { return nil }
func (deleteQueueTask) Action() http.HandlerFunc        { return DeleteQueueTask.action }
func (deleteQueueTask) Page() http.HandlerFunc          { return nil }
func (saveTemplateTask) Action() http.HandlerFunc       { return SaveTemplateTask.action }
func (saveTemplateTask) Page() http.HandlerFunc         { return nil }
func (testTemplateTask) Action() http.HandlerFunc       { return TestTemplateTask.action }
func (testTemplateTask) Page() http.HandlerFunc         { return nil }
func (deleteDLQTask) Action() http.HandlerFunc          { return DeleteDLQTask.action }
func (deleteDLQTask) Page() http.HandlerFunc            { return nil }
func (markReadTask) Action() http.HandlerFunc           { return MarkReadTask.action }
func (markReadTask) Page() http.HandlerFunc             { return nil }
func (purgeNotificationsTask) Action() http.HandlerFunc { return PurgeNotificationsTask.action }
func (purgeNotificationsTask) Page() http.HandlerFunc   { return nil }
func (sendNotificationTask) Action() http.HandlerFunc   { return SendNotificationTask.action }
func (sendNotificationTask) Page() http.HandlerFunc     { return nil }
func (addAnnouncementTask) Action() http.HandlerFunc    { return AddAnnouncementTask.action }
func (addAnnouncementTask) Page() http.HandlerFunc      { return nil }
func (deleteAnnouncementTask) Action() http.HandlerFunc { return DeleteAnnouncementTask.action }
func (deleteAnnouncementTask) Page() http.HandlerFunc   { return nil }
func (addIPBanTask) Action() http.HandlerFunc           { return AddIPBanTask.action }
func (addIPBanTask) Page() http.HandlerFunc             { return nil }
func (deleteIPBanTask) Action() http.HandlerFunc        { return DeleteIPBanTask.action }
func (deleteIPBanTask) Page() http.HandlerFunc          { return nil }
func (newsUserAllowTask) Action() http.HandlerFunc      { return news.NewsAdminUserLevelsAllowActionPage }
func (newsUserAllowTask) Page() http.HandlerFunc        { return nil }
func (newsUserRemoveTask) Action() http.HandlerFunc     { return news.NewsAdminUserLevelsRemoveActionPage }
func (newsUserRemoveTask) Page() http.HandlerFunc       { return nil }

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

var NewsUserAllowTask = newsUserAllowTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskAllow,
		Match:     hcommon.TaskMatcher(hcommon.TaskAllow),
	},
}

var NewsUserRemoveTask = newsUserRemoveTask{
	BasicTaskEvent: eventbus.BasicTaskEvent{
		EventName: hcommon.TaskRemoveLower,
		Match:     hcommon.TaskMatcher(hcommon.TaskRemoveLower),
	},
}
