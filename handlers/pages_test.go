package handlers

import (
	"github.com/arran4/goa4web/core/templates"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHappyPathPagesExist(t *testing.T) {
	pages := []Page{
		TaskErrorAcknowledgementPageTmpl,
		NotFoundPageTmpl,
		AccessDeniedLoginPageTmpl,
		TaskDoneAutoRefreshPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(templates.WithSilence(true)), "Page %s should exist", page)
		})
	}
}
