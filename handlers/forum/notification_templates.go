package forum

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateAdminNotificationForumCategoryCreate notif.EmailTemplateName = "adminNotificationForumCategoryCreateEmail"
	EmailTemplateAdminNotificationForumCategoryUpdate notif.EmailTemplateName = "adminNotificationForumCategoryUpdateEmail"
	EmailTemplateAdminNotificationForumCategoryDelete notif.EmailTemplateName = "adminNotificationForumCategoryDeleteEmail"

	EmailTemplateAdminNotificationForumTopicCreate notif.EmailTemplateName = "adminNotificationForumTopicCreateEmail"
	EmailTemplateAdminNotificationForumTopicChange notif.EmailTemplateName = "adminNotificationForumTopicChangeEmail"
	EmailTemplateAdminNotificationForumTopicDelete notif.EmailTemplateName = "adminNotificationForumTopicDeleteEmail"

	EmailTemplateAdminNotificationForumThreadCreate notif.EmailTemplateName = "adminNotificationForumThreadCreateEmail"
	EmailTemplateAdminNotificationForumThreadDelete notif.EmailTemplateName = "adminNotificationForumThreadDeleteEmail"

	EmailTemplateForumThreadCreate  notif.EmailTemplateName        = "forumThreadCreateEmail"
	NotificationTemplateForumThread notif.NotificationTemplateName = "thread"
)
