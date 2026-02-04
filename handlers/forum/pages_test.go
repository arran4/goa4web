package forum

import (
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []tasks.Template{
		ForumPageTmpl,
		ForumThreadPageTmpl,
		ForumTopicsPageTmpl,
		ForumAdminPageTmpl,
		RunTaskPageTmpl,
		AdminForumFlaggedPostsPageTmpl,
		AdminForumModeratorLogsPageTmpl,
		AdminForumWordListPageTmpl,
		ForumAdminCategoriesPageTmpl,
		ForumAdminCategoryCreatePageTmpl,
		ForumAdminCategoryEditPageTmpl,
		ForumAdminCategoryGrantsPageTmpl,
		ForumAdminCategoryPageTmpl,
		ForumAdminThreadsPageTmpl,
		ConfirmPageTmpl,
		ForumAdminThreadPageTmpl,
		ForumAdminTopicGrantsPageTmpl,
		ForumAdminTopicPageTmpl,
		ForumAdminTopicEditPageTmpl,
		ForumAdminTopicsPageTmpl,
		ForumAdminTopicDeletePageTmpl,
		ForumThreadNewPageTmpl,
		RedirectBackPageTmpl,
		ForumCreateTopicPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(templates.WithSilence(true)), "Page %s should exist", page)
		})
	}
}
