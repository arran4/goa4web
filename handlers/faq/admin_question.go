package faq

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/faq_templates"
	"net/http"
	"strconv"
)

type AdminQuestion struct {
	common.CoreData
	Question *db.Faq
}

type AdminQuestions struct {
	common.CoreData
	Questions           []*db.Faq
	UnansweredQuestions []*db.Faq
	DismissedQuestions  []*db.AdminGetFAQDismissedQuestionsRow
	Category            *db.FaqCategory
	Categories          []*db.FaqCategory
	Templates           []string
}

func (p *AdminQuestions) TemplateName() string {
	return "faq/adminQuestions.gohtml"
}

func (p *AdminQuestion) TemplateName() string {
	return "faq/adminQuestion.gohtml"
}

func (p *AdminQuestions) Load(ctx context.Context, d db.Querier, r *http.Request) error {
	var err error
	cid := r.URL.Query().Get("category")
	if len(cid) > 0 {
		var cat *db.AdminGetFAQCategoryWithQuestionCountByIDRow
		id, _ := strconv.Atoi(cid)
		cat, err = d.AdminGetFAQCategoryWithQuestionCountByID(ctx, int32(id))
		if err == nil {
			p.Category = &db.FaqCategory{
				ID:               cat.ID,
				ParentCategoryID: cat.ParentCategoryID,
				LanguageID:       cat.LanguageID,
				Name:             cat.Name,
			}
			p.Questions, err = d.AdminGetFAQQuestionsByCategory(ctx, sql.NullInt32{Int32: cat.ID, Valid: true})
		}
	} else {
		p.Questions, err = d.AdminGetFAQActiveQuestions(ctx)
	}
	if err != nil {
		return err
	}
	p.Templates, err = faq_templates.List()
	if err != nil {
		return err
	}
	p.Categories, err = d.AdminGetFAQCategories(ctx)
	if err != nil {
		return err
	}
	p.UnansweredQuestions, err = d.AdminGetFAQUnansweredQuestions(ctx)
	if err != nil {
		return err
	}
	p.DismissedQuestions, err = d.AdminGetFAQDismissedQuestions(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *AdminQuestion) Load(ctx context.Context, d db.Querier, r *http.Request) error {
	var err error
	id := r.URL.Query().Get("id")
	if len(id) > 0 {
		val, _ := strconv.Atoi(id)
		p.Question, err = d.AdminGetFAQByID(ctx, int32(val))
	}
	return err
}
