package faq

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
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

	questions, err := queries.SystemGetFAQQuestions(r.Context())
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
