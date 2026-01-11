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
	"github.com/gorilla/mux"
)

// AdminEditQuestionPage displays the edit form for a single FAQ entry.
func AdminEditQuestionPage(w http.ResponseWriter, r *http.Request) {
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
				handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
				return
			}
		}
	} else {
		faq = &db.Faq{ID: 0}
	}
	cats, _ := queries.AdminGetFAQCategories(r.Context())
	type Data struct {
		Faq        *db.Faq
		Categories []*db.FaqCategory
	}
	data := Data{
		Faq:        faq,
		Categories: cats,
	}
	if id != 0 {
		cd.PageTitle = fmt.Sprintf("Edit FAQ %d", id)
	} else {
		cd.PageTitle = "New FAQ"
	}
	AdminQuestionEditPageTmpl.Handle(w, r, data)
}

const AdminQuestionEditPageTmpl handlers.Page = "faq/adminQuestionEditPage.gohtml"

// AdminCreateQuestionPage redirects to AdminEditQuestionPage with id zero to
// display the form for creating a new FAQ entry.
func AdminCreateQuestionPage(w http.ResponseWriter, r *http.Request) {
	r = mux.SetURLVars(r, map[string]string{"id": "0"})
	AdminEditQuestionPage(w, r)
}
