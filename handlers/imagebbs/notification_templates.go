package imagebbs

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateImagePostApproved        notif.EmailTemplateName        = "imagePostApprovedEmail"
	NotificationTemplateImagePostApproved notif.NotificationTemplateName = "image_post_approved"

	EmailTemplateImageBoardUpdate notif.EmailTemplateName = "imageBoardUpdateEmail"

	EmailTemplateAdminNotificationImageBoardNew notif.EmailTemplateName = "adminNotificationImageBoardNewEmail"

	EmailTemplateImagebbsAdminBoard   notif.EmailTemplateName        = "imagebbsAdminBoardEmail"
	NotificationTemplateImagebbsBoard notif.NotificationTemplateName = "imagebbs_board"

	EmailTemplateImagebbsAdminNewBoard   notif.EmailTemplateName        = "imagebbsAdminNewBoardEmail"
	NotificationTemplateImagebbsNewBoard notif.NotificationTemplateName = "imagebbs_new_board"

	EmailTemplateImagebbsReply        notif.EmailTemplateName        = "imagebbsReplyEmail"
	NotificationTemplateImagebbsReply notif.NotificationTemplateName = "imagebbs_reply"
)
