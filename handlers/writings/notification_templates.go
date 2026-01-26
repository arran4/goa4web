package writings

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateWriting        notif.EmailTemplateName        = "writingEmail"
	NotificationTemplateWriting notif.NotificationTemplateName = "writing"

	EmailTemplateWritingUpdate        notif.EmailTemplateName        = "writingUpdateEmail"
	NotificationTemplateWritingUpdate notif.NotificationTemplateName = "writing_update"

	EmailTemplateWritingReply        notif.EmailTemplateName        = "replyEmail"
	NotificationTemplateWritingReply notif.NotificationTemplateName = "reply"

	EmailTemplateAdminNotificationWritingCommentEdit notif.EmailTemplateName = "adminNotificationNewsCommentEditEmail"
)
