package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// RemakeThreadStatsTask refreshes forum thread statistics.
type RemakeThreadStatsTask struct{ tasks.TaskString }

var remakeThreadStatsTask = &RemakeThreadStatsTask{TaskString: TaskRemakeStatisticInformationOnForumthread}

// RemakeTopicStatsTask refreshes forum topic statistics.
type RemakeTopicStatsTask struct{ tasks.TaskString }

var remakeTopicStatsTask = &RemakeTopicStatsTask{TaskString: TaskRemakeStatisticInformationOnForumtopic}

// CategoryChangeTask updates a forum category name.
type CategoryChangeTask struct{ tasks.TaskString }

var categoryChangeTask = &CategoryChangeTask{TaskString: TaskForumCategoryChange}

// CategoryCreateTask creates a new forum category.
type CategoryCreateTask struct{ tasks.TaskString }

var categoryCreateTask = &CategoryCreateTask{TaskString: TaskForumCategoryCreate}

// DeleteCategoryTask removes a forum category.
type DeleteCategoryTask struct{ tasks.TaskString }

var deleteCategoryTask = &DeleteCategoryTask{TaskString: TaskDeleteCategory}

// ThreadDeleteTask removes a forum thread.
type ThreadDeleteTask struct{ tasks.TaskString }

var threadDeleteTask = &ThreadDeleteTask{TaskString: TaskForumThreadDelete}

var (
	_ tasks.Task                       = (*CategoryChangeTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CategoryChangeTask)(nil)
	_ tasks.Task                       = (*CategoryCreateTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*CategoryCreateTask)(nil)
	_ tasks.Task                       = (*DeleteCategoryTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*DeleteCategoryTask)(nil)
	_ tasks.Task                       = (*ThreadDeleteTask)(nil)
	_ notif.AdminEmailTemplateProvider = (*ThreadDeleteTask)(nil)
)

func (CategoryChangeTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumCategoryChangeEmail")
}

func (CategoryChangeTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumCategoryChangeEmail")
	return &v
}

func (CategoryCreateTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumCategoryCreateEmail")
}

func (CategoryCreateTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumCategoryCreateEmail")
	return &v
}

func (DeleteCategoryTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumDeleteCategoryEmail")
}

func (DeleteCategoryTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumDeleteCategoryEmail")
	return &v
}

func (ThreadDeleteTask) AdminEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("adminNotificationForumThreadDeleteEmail")
}

func (ThreadDeleteTask) AdminInternalNotificationTemplate() *string {
	v := notif.NotificationTemplateFilenameGenerator("adminNotificationForumThreadDeleteEmail")
	return &v
}
