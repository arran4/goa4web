package faq

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/faq_templates"
)

// AdminCategoryPage displays information about a FAQ category including recent questions.
func AdminCategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Category  *db.AdminGetFAQCategoryWithQuestionCountByIDRow
		Latest    []*db.Faq
		Templates []string
		Grants    []*db.SearchGrantsRow
		Roles     []*db.Role
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid category id"))
		return
	}

	queries := cd.Queries()

	cat, err := queries.AdminGetFAQCategoryWithQuestionCountByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	latest, err := queries.AdminGetFAQQuestionsByCategory(r.Context(), sql.NullInt32{Int32: int32(id), Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	templates, err := faq_templates.List()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	grants, err := queries.SearchGrants(r.Context(), db.SearchGrantsParams{
		Section: sql.NullString{String: "faq", Valid: true},
		Item:    sql.NullString{String: "category", Valid: true},
		ItemID:  sql.NullInt32{Int32: int32(id), Valid: true},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}

	roles, err := cd.AllRoles()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	cd.PageTitle = "FAQ Category: " + cat.Name.String

	data := Data{Category: cat, Latest: latest, Templates: templates, Grants: grants, Roles: roles}
	FaqAdminCategoryPageTmpl.Handle(w, r, data)
}

const FaqAdminCategoryPageTmpl tasks.Template = "faq/faqAdminCategoryPage.gohtml"

// AdminCategoryEditPage shows a form to rename or delete a FAQ category.
func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Category   *db.AdminGetFAQCategoryWithQuestionCountByIDRow
		Categories []*db.FaqCategory
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid category id"))
		return
	}

	queries := cd.Queries()
	cat, err := queries.AdminGetFAQCategoryWithQuestionCountByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	cats, err := queries.AdminGetFAQCategories(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	cd.PageTitle = "Edit FAQ Category"
	data := Data{Category: cat, Categories: cats}
	FaqAdminCategoryEditPageTmpl.Handle(w, r, data)
}

const FaqAdminCategoryEditPageTmpl tasks.Template = "faq/faqAdminCategoryEditPage.gohtml"

// AdminCategoryQuestionsPage lists questions for a FAQ category.
func AdminCategoryQuestionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Category  *db.AdminGetFAQCategoryWithQuestionCountByIDRow
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
	cat, err := queries.AdminGetFAQCategoryWithQuestionCountByID(r.Context(), int32(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			handlers.RenderErrorPage(w, r, fmt.Errorf("category not found"))
			return
		default:
			handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
			return
		}
	}

	questions, err := queries.AdminGetFAQQuestionsByCategory(r.Context(), sql.NullInt32{Int32: int32(id), Valid: true})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	cd.PageTitle = "FAQ Category Questions"
	data := Data{Category: cat, Questions: questions}
	FaqAdminCategoryQuestionsPageTmpl.Handle(w, r, data)
}

const FaqAdminCategoryQuestionsPageTmpl tasks.Template = "faq/faqAdminCategoryQuestionsPage.gohtml"

// AdminNewCategoryPage displays a form to create a new FAQ category.
func AdminNewCategoryPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "New FAQ Category"
	FaqAdminNewCategoryPageTmpl.Handle(w, r, struct{}{})
}

const FaqAdminNewCategoryPageTmpl tasks.Template = "faq/faqAdminNewCategoryPage.gohtml"
