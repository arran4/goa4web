package faq

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestHappyPathAdminQuestions_Load(t *testing.T) {
	qs := testhelpers.NewQuerierStub()
	qs.AdminGetFAQActiveQuestionsFn = func(ctx context.Context) ([]*db.Faq, error) {
		return []*db.Faq{
			{ID: 1, Question: sql.NullString{String: "Q1", Valid: true}, Answer: sql.NullString{String: "A1", Valid: true}},
		}, nil
	}
	qs.AdminGetFAQCategoriesFn = func(ctx context.Context) ([]*db.FaqCategory, error) {
		return []*db.FaqCategory{}, nil
	}
	qs.AdminGetFAQUnansweredQuestionsFn = func(ctx context.Context) ([]*db.Faq, error) {
		return []*db.Faq{}, nil
	}
	qs.AdminGetFAQDismissedQuestionsFn = func(ctx context.Context) ([]*db.AdminGetFAQDismissedQuestionsRow, error) {
		return []*db.AdminGetFAQDismissedQuestionsRow{}, nil
	}

	p := &AdminQuestions{
		CoreData: common.CoreData{},
	}
	r := httptest.NewRequest("GET", "/admin/faq/questions", nil)

	err := p.Load(context.Background(), qs, r)
	assert.NoError(t, err)

	t.Run("Page Content Verification", func(t *testing.T) {
		assert.Len(t, p.Questions, 1)
		assert.Equal(t, "Q1", p.Questions[0].Question.String)
		assert.NotEmpty(t, p.Templates)
	})
}
