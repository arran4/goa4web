package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

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
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	var category *db.FaqCategory
	if faq.CategoryID.Valid {
		category, err = queries.AdminGetFAQCategory(r.Context(), faq.CategoryID.Int32)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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

const AdminQuestionPageTmpl tasks.Template = "faq/adminQuestionPage.gohtml"
