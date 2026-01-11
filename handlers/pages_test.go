package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagesExist(t *testing.T) {
	pages := []Page{
		TaskErrorAcknowledgementPageTmpl,
		NotFoundPageTmpl,
		AccessDeniedLoginPageTmpl,
		TaskDoneAutoRefreshPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(), "Page %s should exist", page)
		})
	}
}
