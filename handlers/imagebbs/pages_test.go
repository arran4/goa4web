package imagebbs

import (
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []tasks.Template{
		ImageBBSBoardThreadPageTmpl,
		ImageBBSAdminBoardPageTmpl,
		ImageBBSAdminPostEditPageTmpl,
		ImageBBSAdminPostDashboardPageTmpl,
		ImageBBSAdminPostCommentsPageTmpl,
		ImageBBSAdminPageTmpl,
		ImageBBSAdminBoardViewPageTmpl,
		ImageBBSAdminBoardsPageTmpl,
		ImageBBSPosterPageTmpl,
		ImageBBSAdminBoardListPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(templates.WithSilence(true)), "Page %s should exist", page)
		})
	}
}
