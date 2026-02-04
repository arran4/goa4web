package faq

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminQuestionsPage renders the questions administration view.
func AdminQuestionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	unanswered, err := queries.AdminGetFAQUnansweredQuestions(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	var questions []*db.Faq
	catID := r.URL.Query().Get("category")
	if catID != "" {
		if catID == "0" {
			questions, err = queries.AdminGetFAQUnansweredQuestions(r.Context())
		} else {
			cid, _ := strconv.Atoi(catID)
			questions, err = queries.AdminGetFAQQuestionsByCategory(r.Context(), sql.NullInt32{Int32: int32(cid), Valid: true})
		}
	} else {
		questions, err = queries.SystemGetFAQQuestions(r.Context())
	}
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	dismissed, err := queries.AdminGetFAQDismissedQuestions(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	cats, err := queries.AdminGetFAQCategories(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	cd.PageTitle = "FAQ Questions"

	data := map[string]any{
		"UnansweredQuestions": unanswered,
		"Questions":           questions,
		"DismissedQuestions":  dismissed,
		"Categories":          cats,
	}

	AdminQuestionsPageTmpl.Handle(w, r, data)
}

const AdminQuestionsPageTmpl tasks.Template = "faq/adminQuestions.gohtml"
