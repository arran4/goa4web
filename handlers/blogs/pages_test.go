package blogs

import (
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []handlers.Page{
		BlogsPageTmpl,
		BlogsBlogPageTmpl,
		BlogsBlogEditPageTmpl,
		BlogsBlogAddPageTmpl,
		BlogsBloggersBloggerPageTmpl,
		BloggerPostsPageTmpl,
		BloggerListPageTmpl,
		BlogsAdminBlogCommentsPageTmpl,
		BlogsAdminBlogEditPageTmpl,
		BlogsAdminBlogPageTmpl,
		BlogsAdminPageTmpl,
		BlogsCommentPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(), "Page %s should exist", page)
		})
	}
}
