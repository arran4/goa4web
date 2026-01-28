package blogs

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateAdminNotificationBlogAdd notif.EmailTemplateName        = "adminNotificationBlogAddEmail"
	EmailTemplateBlogAdd                  notif.EmailTemplateName        = "blogAddEmail"
	NotificationTemplateBlogAdd           notif.NotificationTemplateName = "blog_add"

	EmailTemplateAdminNotificationBlogEdit notif.EmailTemplateName        = "adminNotificationBlogEditEmail"
	EmailTemplateBlogEdit                  notif.EmailTemplateName        = "blogEditEmail"
	NotificationTemplateBlogEdit           notif.NotificationTemplateName = "blog_edit"

	EmailTemplateBlogReply        notif.EmailTemplateName        = "replyEmail"
	NotificationTemplateBlogReply notif.NotificationTemplateName = "reply"

	EmailTemplateAdminNotificationBlogCommentEdit   notif.EmailTemplateName = "adminNotificationBlogCommentEditEmail"
	EmailTemplateAdminNotificationBlogCommentCancel notif.EmailTemplateName = "adminNotificationBlogCommentCancelEmail"
)
