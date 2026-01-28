package search

import (
	notif "github.com/arran4/goa4web/internal/notifications"
)

const (
	EmailTemplateSearchRebuildBlog        notif.EmailTemplateName        = "searchRebuildBlogEmail"
	NotificationTemplateSearchRebuildBlog notif.NotificationTemplateName = "search_rebuild_blog"

	EmailTemplateSearchRebuildComments        notif.EmailTemplateName        = "searchRebuildCommentsEmail"
	NotificationTemplateSearchRebuildComments notif.NotificationTemplateName = "search_rebuild_comments"

	EmailTemplateSearchRebuildImage        notif.EmailTemplateName        = "searchRebuildImageEmail"
	NotificationTemplateSearchRebuildImage notif.NotificationTemplateName = "search_rebuild_image"

	EmailTemplateSearchRebuildLinker        notif.EmailTemplateName        = "searchRebuildLinkerEmail"
	NotificationTemplateSearchRebuildLinker notif.NotificationTemplateName = "search_rebuild_linker"

	EmailTemplateSearchRebuildNews        notif.EmailTemplateName        = "searchRebuildNewsEmail"
	NotificationTemplateSearchRebuildNews notif.NotificationTemplateName = "search_rebuild_news"

	EmailTemplateSearchRebuildWriting        notif.EmailTemplateName        = "searchRebuildWritingEmail"
	NotificationTemplateSearchRebuildWriting notif.NotificationTemplateName = "search_rebuild_writing"
)
