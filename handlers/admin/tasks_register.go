package admin

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns admin related tasks.
func RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		addAnnouncementTask,
		deleteAnnouncementTask,
		resendQueueTask,
		deleteQueueTask,
		saveTemplateTask,
		testTemplateTask,
		deleteDLQTask,
		markReadTask,
		purgeSelectedNotificationsTask,
		purgeReadNotificationsTask,
		toggleNotificationReadTask,
		sendNotificationTask,
		addIPBanTask,
		deleteIPBanTask,
		acceptRequestTask,
		rejectRequestTask,
		queryRequestTask,
		newsUserAllow,
		newsUserRemove,
	}
}
