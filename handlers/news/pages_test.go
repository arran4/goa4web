package news_test

import (
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
	"testing"

	"github.com/arran4/goa4web/handlers/news"
)

var allPages = []tasks.Template{
	news.AdminNewsListPageTmpl,
	news.AdminNewsPostPageTmpl,
	news.AdminNewsEditPageTmpl,
	news.AdminNewsDeleteConfirmPageTmpl,
	news.NewsCreatePageTmpl,
	news.NewsPostPageTmpl,
	news.NewsEditPageTmpl,
	news.NewsPreviewPageTmpl,
	news.NewsPageTmpl,
	news.SearchResultNewsActionPageTmpl,
}

func TestHappyPathAllRegisteredPagesExist(t *testing.T) {
	for _, p := range allPages {
		if !p.Exists(templates.WithSilence(true)) {
			t.Errorf("Page template missing: %s", p)
		}
	}
}
