package forum

import (
	"testing"

	coretemplates "github.com/arran4/goa4web/core/templates"
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

	dir := "../../core/templates"
	for _, tmpl := range pageTemplates {
		if !coretemplates.IsTemplateAvailable(tmpl, dir) {
			t.Errorf("Template %s not found", tmpl)
		}
	}
}
