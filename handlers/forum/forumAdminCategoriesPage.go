package forum

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Categories []*GetAllForumCategoriesWithSubcategoryCountRow
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	categoryRows, err := queries.GetAllForumCategoriesWithSubcategoryCount(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllForumCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	CustomForumIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminCategoriesPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminCategoryEditPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	vars := mux.Vars(r)
	categoryId, _ := strconv.Atoi(vars["category"])

	if err := queries.UpdateForumCategory(r.Context(), UpdateForumCategoryParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idforumcategory:              int32(categoryId),
		ForumcategoryIdforumcategory: int32(pcid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/forum/admin/categories", http.StatusTemporaryRedirect)
}

func AdminCategoryCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(common.KeyQueries).(*Queries)
	if err := queries.CreateForumCategory(r.Context(), CreateForumCategoryParams{
		ForumcategoryIdforumcategory: int32(pcid),
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, "/forum/admin/categories", http.StatusTemporaryRedirect)
}

func AdminCategoryDeletePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*Queries)
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if err := queries.DeleteForumCategory(r.Context(), int32(cid)); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/forum/admin/categories", http.StatusTemporaryRedirect)
}
