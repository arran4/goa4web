package faq

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestAdminQuestions_Load(t *testing.T) {
	qs := &db.QuerierStub{
		AdminGetFAQActiveQuestionsFn: func(ctx context.Context) ([]*db.Faq, error) {
			return []*db.Faq{
				{ID: 1, Question: sql.NullString{String: "Q1", Valid: true}, Answer: sql.NullString{String: "A1", Valid: true}},
			}, nil
		},
		AdminGetFAQCategoriesFn: func(ctx context.Context) ([]*db.FaqCategory, error) {
			return []*db.FaqCategory{}, nil
		},
		AdminGetFAQUnansweredQuestionsFn: func(ctx context.Context) ([]*db.Faq, error) {
			return []*db.Faq{}, nil
		},
		AdminGetFAQDismissedQuestionsFn: func(ctx context.Context) ([]*db.AdminGetFAQDismissedQuestionsRow, error) {
			return []*db.AdminGetFAQDismissedQuestionsRow{}, nil
		},
	}

	p := &AdminQuestions{
		CoreData: common.CoreData{},
	}
	r := httptest.NewRequest("GET", "/admin/faq/questions", nil)

	err := p.Load(context.Background(), qs, r)
	assert.NoError(t, err)
	assert.Len(t, p.Questions, 1)
	assert.Equal(t, "Q1", p.Questions[0].Question.String)
	assert.NotEmpty(t, p.Templates) // Assuming at least one template exists in embedded FS
}
