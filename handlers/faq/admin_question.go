package faq

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
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
}

func (p *AdminQuestions) TemplateName() string {
	return "faq/adminQuestions.gohtml"
}

func (p *AdminQuestion) TemplateName() string {
	return "faq/adminQuestion.gohtml"
}

func (p *AdminQuestions) Load(ctx context.Context, d *db.Queries, r *http.Request) error {
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
		var answeredRows []*db.GetFAQAnsweredQuestionsRow
		answeredRows, err = d.GetFAQAnsweredQuestions(ctx, db.GetFAQAnsweredQuestionsParams{
			ViewerID: p.UserID,
			UserID:   sql.NullInt32{Int32: p.UserID, Valid: true},
		})
		if err == nil {
			p.Questions = make([]*db.Faq, len(answeredRows))
			for i, r := range answeredRows {
				p.Questions[i] = &db.Faq{
					ID:         r.ID,
					CategoryID: r.CategoryID,
					LanguageID: r.LanguageID,
					AuthorID:   r.AuthorID,
					Answer:     r.Answer,
					Question:   r.Question,
					// Priority missing from row, defaulted to 0
				}
			}
		}
	}
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

func (p *AdminQuestion) Load(ctx context.Context, d *db.Queries, r *http.Request) error {
	var err error
	id := r.URL.Query().Get("id")
	if len(id) > 0 {
		val, _ := strconv.Atoi(id)
		p.Question, err = d.AdminGetFAQByID(ctx, int32(val))
	}
	return err
}
