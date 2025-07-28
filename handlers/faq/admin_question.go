package faq

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminQuestionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Categories []*db.FaqCategory
		Rows       []*db.Faq
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	cd := data.CoreData
	cd.PageTitle = "FAQ Questions"

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	catrows, err := queries.GetAllFAQCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Categories = catrows

	rows, err := queries.GetAllFAQQuestions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	handlers.TemplateHandler(w, r, "adminQuestionPage.gohtml", data)
}
