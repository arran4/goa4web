package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminQuestionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories     []*db.FaqCategory
		UnansweredRows []*db.Faq
		AnsweredRows   []*db.Faq
		DismissedRows  []*db.Faq
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	cd := data.CoreData

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	catrows, err := queries.GetAllFAQCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data.Categories = catrows

	cd.PageTitle = "FAQ Questions"

	unansweredRows, err := queries.GetFAQUnansweredQuestions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data.UnansweredRows = unansweredRows

	answeredRows, err := queries.GetFAQAnsweredQuestions(r.Context(), db.GetFAQAnsweredQuestionsParams{
		ViewerID: cd.UserID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data.AnsweredRows = answeredRows

	dismissedRows, err := queries.GetFAQDismissedQuestions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}
	data.DismissedRows = dismissedRows

	handlers.TemplateHandler(w, r, "adminQuestionPage.gohtml", data)
}
