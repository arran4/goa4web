package faq

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows []*db.AdminGetFAQCategoriesWithQuestionCountRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "FAQ Categories"
	data := Data{}

	queries := cd.Queries()

	rows, err := queries.AdminGetFAQCategoriesWithQuestionCount(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}
	data.Rows = rows

	FaqAdminCategoriesPageTmpl.Handle(w, r, data)
}

const FaqAdminCategoriesPageTmpl tasks.Template = "faq/faqAdminCategoriesPage.gohtml"
