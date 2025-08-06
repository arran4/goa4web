package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminCategoryPage displays information about a FAQ category including recent questions.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Category *db.GetFAQCategoryWithQuestionCountByIDRow
		Latest   []*db.Faq
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid category id"))
		return
	}

	queries := cd.Queries()

	cat, err := queries.GetFAQCategoryWithQuestionCountByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	latest, err := queries.GetFAQQuestionsByCategory(r.Context(), int32(id))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	if len(latest) > 5 {
		latest = latest[:5]
	}

	cd.PageTitle = "FAQ Category"

	data := Data{Category: cat, Latest: latest}
	handlers.TemplateHandler(w, r, "faqAdminCategoryPage.gohtml", data)
}

// AdminCategoryEditPage shows a form to rename or delete a FAQ category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Category *db.GetFAQCategoryWithQuestionCountByIDRow
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid category id"))
		return
	}

	queries := cd.Queries()
	cat, err := queries.GetFAQCategoryWithQuestionCountByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	cd.PageTitle = "Edit FAQ Category"
	data := Data{Category: cat}
	handlers.TemplateHandler(w, r, "faqAdminCategoryEditPage.gohtml", data)
}

// AdminCategoryQuestionsPage lists questions for a FAQ category.
func AdminCategoryQuestionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Category  *db.GetFAQCategoryWithQuestionCountByIDRow
		Questions []*db.Faq
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid category id"))
		return
	}

	queries := cd.Queries()
	cat, err := queries.GetFAQCategoryWithQuestionCountByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	questions, err := queries.GetFAQQuestionsByCategory(r.Context(), int32(id))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	cd.PageTitle = "FAQ Category Questions"
	data := Data{Category: cat, Questions: questions}
	handlers.TemplateHandler(w, r, "faqAdminCategoryQuestionsPage.gohtml", data)
}

// AdminNewCategoryPage displays a form to create a new FAQ category.
func AdminNewCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "New FAQ Category"
	handlers.TemplateHandler(w, r, "faqAdminNewCategoryPage.gohtml", struct{}{})
}
