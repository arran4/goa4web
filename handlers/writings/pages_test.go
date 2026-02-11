package writings_test

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/writings"
	"github.com/arran4/goa4web/internal/tasks"
)

var allPages = []tasks.Template{
	writings.ArticlePageTmpl,
	writings.AdminNoAccessPageTmpl,
	writings.WritingsPageTmpl,
	writings.WritingsCategoryPageTmpl,
	writings.WritingsArticleAddPageTmpl,
	writings.WritingsArticleEditPageTmpl,
	writings.WritingsCategoriesPageTmpl,
	writings.WritingsWriterListPageTmpl,
	writings.WritingsWriterPageTmpl,
	writings.AdminWritingsPageTmpl,
	writings.WritingsAdminCategoriesPageTmpl,
	writings.WritingsAdminCategoryEditPageTmpl,
	writings.WritingsAdminCategoryGrantsPageTmpl,
	writings.WritingsAdminCategoryPageTmpl,
	writings.WritingsAdminPageTmpl,
}

func TestAllRegisteredPagesExist(t *testing.T) {
	t.Run("Happy Path - All Pages", func(t *testing.T) {
		for _, p := range allPages {
			if !p.Exists(templates.WithSilence(true)) {
				t.Errorf("Page template missing: %s", p)
			}
		}
	})
}
