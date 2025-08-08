package admin

import "github.com/arran4/goa4web/internal/tasks"

// RegisterTasks returns admin related tasks.
func (h *Handlers) RegisterTasks() []tasks.NamedTask {
	return []tasks.NamedTask{
		addAnnouncementTask,
		deleteAnnouncementTask,
		resendQueueTask,
		deleteQueueTask,
		resendSentEmailTask,
		retrySentEmailTask,
		saveTemplateTask,
		deleteTemplateTask,
		testTemplateTask,
		deleteDLQTask,
		markReadTask,
		markUnreadTask,
		toggleNotificationReadTask,
		purgeSelectedNotificationsTask,
		purgeReadNotificationsTask,
		deleteNotificationTask,
		sendNotificationTask,
		addUserSubscriptionTask,
		updateUserSubscriptionTask,
		deleteUserSubscriptionTask,
		addIPBanTask,
		deleteIPBanTask,
		acceptRequestTask,
		rejectRequestTask,
		queryRequestTask,
		deleteCommentTask,
		editCommentTask,
		banCommentTask,
		userPasswordResetTask,
		roleGrantCreateTask,
		roleGrantDeleteTask,
		userGrantCreateTask,
		applyLinkRemapTask,
		h.NewServerShutdownTask(),
	}
}
