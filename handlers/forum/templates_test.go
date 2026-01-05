package forum

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"
)

func TestTemplatesExist(t *testing.T) {
	pageTemplates := []string{
		ForumAdminFlaggedPostsPageTmpl,
		ForumAdminModeratorLogsPageTmpl,
		ForumAdminWordListPageTmpl,
		ForumAdminCategoriesPageTmpl,
		ForumAdminCategoryCreatePageTmpl,
		ForumAdminCategoryEditPageTmpl,
		ForumAdminCategoryGrantsPageTmpl,
		ForumAdminCategoryPageTmpl,
		ForumAdminThreadsPageTmpl,
		ForumAdminThreadPageTmpl,
		ForumAdminTopicGrantsPageTmpl,
		ForumAdminTopicPageTmpl,
		ForumAdminTopicEditPageTmpl,
		ForumAdminTopicsPageTmpl,
		ForumAdminTopicDeletePageTmpl,
		ForumPageTmpl,
		ForumThreadNewPageTmpl,
		ForumThreadPageTmpl,
		ForumTopicPageTmpl,
		ForumCreateTopicPageTmpl,
		RedirectBackPageTmpl,
	}

	templates.SetDir("../../core/templates")
	for _, tmpl := range pageTemplates {
		if !templates.IsTemplateAvailable(tmpl) {
			t.Errorf("Template %s not found", tmpl)
		}
	}
}
