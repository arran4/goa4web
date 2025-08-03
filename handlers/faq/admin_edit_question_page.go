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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.SetCurrentFAQ(int32(id))
	if id != 0 {
		if _, err := cd.FAQByID(int32(id)); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			default:
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
	}
	if _, err := cd.FAQCategories(); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	if id != 0 {
		cd.PageTitle = fmt.Sprintf("Edit FAQ %d", id)
	} else {
		cd.PageTitle = "New FAQ"
	}
	handlers.TemplateHandler(w, r, "adminQuestionEditPage.gohtml", struct{}{})
}

// AdminCreateQuestionPage redirects to AdminEditQuestionPage with id zero to
// display the form for creating a new FAQ entry.
func AdminCreateQuestionPage(w http.ResponseWriter, r *http.Request) {
	r = mux.SetURLVars(r, map[string]string{"id": "0"})
	AdminEditQuestionPage(w, r)
}
