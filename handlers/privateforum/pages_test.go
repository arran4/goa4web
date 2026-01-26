package privateforum

import (
	"github.com/arran4/goa4web/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []tasks.Template{
		CreateTopicTmpl,
		TopicsOnlyTmpl,
		SharedPreviewLoginPageTmpl,
		PrivateForumStartDiscussionPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(), "Page %s should exist", page)
		})
	}
}
