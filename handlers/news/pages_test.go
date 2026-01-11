package news_test

import (
	"testing"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/news"
)

var allPages = []handlers.Page{
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

func TestAllRegisteredPagesExist(t *testing.T) {
	for _, p := range allPages {
		if !p.Exists() {
			t.Errorf("Page template missing: %s", p)
		}
	}
}
