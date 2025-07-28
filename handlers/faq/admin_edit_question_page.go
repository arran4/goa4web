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
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	var faq *db.Faq
	if id != 0 {
		var err error
		faq, err = queries.GetFAQByID(r.Context(), int32(id))
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			default:
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		faq = &db.Faq{Idfaq: 0}
	}
	cats, _ := queries.GetAllFAQCategories(r.Context())
	type Data struct {
		*common.CoreData
		Faq        *db.Faq
		Categories []*db.FaqCategory
	}
	data := Data{
		CoreData:   r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Faq:        faq,
		Categories: cats,
	}
	cd := data.CoreData
	if id != 0 {
		cd.PageTitle = fmt.Sprintf("Edit FAQ %d", id)
	} else {
		cd.PageTitle = "New FAQ"
	}
	handlers.TemplateHandler(w, r, "adminQuestionEditPage.gohtml", data)
}
