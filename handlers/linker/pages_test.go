package linker

import (
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []tasks.Template{
		LinkerShowPageTmpl,
		LinkerCommentsPageTmpl,
		LinkerAdminDashboardPageTmpl,
		LinkerSuggestPageTmpl,
		LinkerAdminAddPageTmpl,
		LinkerUserPageTmpl,
		LinkerPageTmpl,
		LinkerCategoriesPageTmpl,
		LinkerAdminQueuePageTmpl,
		LinkerAdminCategoriesPageTmpl,
		LinkerAdminCategoryGrantsPageTmpl,
		LinkerAdminLinkPageTmpl,
		LinkerAdminLinkGrantsPageTmpl,
		LinkerAdminCategoryEditPageTmpl,
		LinkerAdminLinkViewPageTmpl,
		LinkerAdminLinksPageTmpl,
		LinkerAdminCategoryPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(templates.WithSilence(true)), "Page %s should exist", page)
		})
	}
}
