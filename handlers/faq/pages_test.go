package faq

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/stretchr/testify/assert"
)

func TestHappyPathPagesExist(t *testing.T) {
	pages := []tasks.Template{
		FaqPageTmpl,
		AskPageTmpl,
		FaqAdminCategoriesPageTmpl,
		FaqAdminCategoryPageTmpl,
		FaqAdminCategoryEditPageTmpl,
		FaqAdminCategoryQuestionsPageTmpl,
		FaqAdminNewCategoryPageTmpl,
		AdminQuestionEditPageTmpl,
		AdminFaqRevisionPageTmpl,
	}

	for _, page := range pages {
		t.Run(string(page), func(t *testing.T) {
			assert.True(t, page.Exists(templates.WithSilence(true)), "Page %s should exist", page)
		})
	}
}
