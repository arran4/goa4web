package linker

import (
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []handlers.Page{
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
			assert.True(t, page.Exists(), "Page %s should exist", page)
		})
	}
}
