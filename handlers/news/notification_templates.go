package news

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateAdminNotificationNewsAdd notif.EmailTemplateName        = "adminNotificationNewsAddEmail"
	EmailTemplateNewsAdd                  notif.EmailTemplateName        = "newsAddEmail"
	NotificationTemplateNewsAdd           notif.NotificationTemplateName = "news_add"

	EmailTemplateAdminNotificationNewsReply notif.EmailTemplateName        = "adminNotificationNewsReplyEmail"
	EmailTemplateNewsReply                  notif.EmailTemplateName        = "replyEmail"
	NotificationTemplateNewsReply           notif.NotificationTemplateName = "reply"

	EmailTemplateAdminNotificationNewsEdit notif.EmailTemplateName        = "adminNotificationNewsEditEmail"
	EmailTemplateNewsEdit                  notif.EmailTemplateName        = "newsEditEmail"
	NotificationTemplateNewsEdit           notif.NotificationTemplateName = "news_edit"

	EmailTemplateAdminNotificationNewsUserAllow     notif.EmailTemplateName = "adminNotificationNewsUserAllowEmail"
	EmailTemplateAdminNotificationNewsUserDisallow  notif.EmailTemplateName = "adminNotificationNewsUserDisallowEmail"
	EmailTemplateAdminNotificationNewsDelete        notif.EmailTemplateName = "adminNotificationNewsDeleteEmail"
	EmailTemplateAdminNotificationNewsCommentEdit   notif.EmailTemplateName = "adminNotificationNewsCommentEditEmail"
	EmailTemplateAdminNotificationNewsCommentCancel notif.EmailTemplateName = "adminNotificationNewsCommentCancelEmail"
)
