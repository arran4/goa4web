package faq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/faq_templates"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AdminQuestions struct {
	common.CoreData
	Questions           []*db.Faq
	UnansweredQuestions []*db.Faq
	DismissedQuestions  []*db.AdminGetFAQDismissedQuestionsRow
	Category            *db.FaqCategory
	Categories          []*db.FaqCategory
	Templates           []string
}

type AdminQuestion struct {
	common.CoreData
	Question *db.Faq
}

// AdminQuestionPage displays a single FAQ question.
func AdminQuestionPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Faq      *db.Faq
		Category *db.FaqCategory
		Author   *db.SystemGetUserByIDRow
		Language string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid question id"))
		return
	}

	queries := cd.Queries()
	faq, err := queries.AdminGetFAQByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("question not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}
	var category *db.FaqCategory
	if faq.CategoryID.Valid {
		category, err = queries.AdminGetFAQCategory(r.Context(), faq.CategoryID.Int32)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	author := cd.UserByID(faq.AuthorID)

	var languageName string
	if faq.LanguageID.Valid {
		if langs, err := cd.Languages(); err == nil {
			for _, l := range langs {
				if l.ID == faq.LanguageID.Int32 {
					languageName = l.Nameof.String
					break
				}
			}
		}
	}

	cd.PageTitle = fmt.Sprintf("FAQ: %s", faq.Question.String)
	data := Data{Faq: faq, Category: category, Author: author, Language: languageName}
	AdminQuestionPageTmpl.Handle(w, r, data)
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

const AdminQuestionPageTmpl tasks.Template = "faq/adminQuestionPage.gohtml"
