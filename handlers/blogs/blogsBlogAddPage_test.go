package blogs

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/notifications"
)

func requireEmailTemplates(t *testing.T, prefix string) {
	t.Helper()
	htmlTmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	textTmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	if htmlTmpls.Lookup(notifications.EmailHTMLTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing html template %s.gohtml", prefix)
	}
	if textTmpls.Lookup(notifications.EmailTextTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing text template %s.gotxt", prefix)
	}
	if textTmpls.Lookup(notifications.EmailSubjectTemplateFilenameGenerator(prefix)) == nil {
		t.Errorf("missing subject template %sSubject.gotxt", prefix)
	}
}

func TestBlogTemplatesExist(t *testing.T) {
	prefixes := []string{
		"blogAddEmail",
		"adminNotificationBlogAddEmail",
		"blogEditEmail",
		"adminNotificationBlogEditEmail",
		"blogReplyEmail",
		"adminNotificationBlogReplyEmail",
		"blogCommentEditEmail",
		"adminNotificationBlogCommentEditEmail",
		"blogCommentCancelEmail",
		"adminNotificationBlogCommentCancelEmail",
		"blogUserAllowEmail",
		"adminNotificationBlogUserAllowEmail",
		"blogUserDisallowEmail",
		"adminNotificationBlogUserDisallowEmail",
		"blogUsersAllowEmail",
		"adminNotificationBlogUsersAllowEmail",
		"blogUsersDisallowEmail",
		"adminNotificationBlogUsersDisallowEmail",
	}
	for _, p := range prefixes {
		requireEmailTemplates(t, p)
	}
}
