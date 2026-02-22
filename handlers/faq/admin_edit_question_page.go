package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/faq_templates"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

type AdminQuestionEditPageTask struct {
}

func (t *AdminQuestionEditPageTask) Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}
	queries := cd.Queries()
	var faq *db.Faq
	if id != 0 {
		var err error
		faq, err = queries.AdminGetFAQByID(r.Context(), int32(id))
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				handlers.RenderErrorPage(w, r, handlers.ErrNotFound)
				return
			default:
				handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
				return
			}
		}
	} else {
		faq = &db.Faq{ID: 0}
	}
	cats, _ := queries.AdminGetFAQCategories(r.Context())
	type TemplateContent struct {
		Question    string
		Answer      string
		Description string
	}
	type Data struct {
		Faq          *db.Faq
		Categories   []*db.FaqCategory
		Templates    []string
		TemplateData map[string]TemplateContent
	}
	templates, _ := faq_templates.List()
	templateData := make(map[string]TemplateContent)
	for _, t := range templates {
		content, err := faq_templates.Get(t)
		if err == nil {
			desc, q, a, err := faq_templates.ParseTemplateContent(content)
			if err == nil {
				templateData[t] = TemplateContent{
					Question:    q,
					Answer:      a,
					Description: desc,
				}
			}
		}
	}
	data := Data{
		Faq:          faq,
		Categories:   cats,
		Templates:    templates,
		TemplateData: templateData,
	}
	if id != 0 {
		cd.PageTitle = fmt.Sprintf("Edit FAQ %d", id)
	} else {
		cd.PageTitle = "New FAQ"
	}
	AdminQuestionEditPageTmpl.Handle(w, r, data)
}

const AdminQuestionEditPageTmpl tasks.Template = "faq/adminQuestionEditPage.gohtml"

// AdminCreateQuestionPage redirects to AdminEditQuestionPage with id zero to
// display the form for creating a new FAQ entry.
func AdminCreateQuestionPage(w http.ResponseWriter, r *http.Request) {
	r = mux.SetURLVars(r, map[string]string{"id": "0"})
	(&AdminQuestionEditPageTask{}).Page(w, r)
}

// AdminEditQuestionPage displays the edit form for a single FAQ entry.
func AdminEditQuestionPage(w http.ResponseWriter, r *http.Request) {
	(&AdminQuestionEditPageTask{}).Page(w, r)
}
