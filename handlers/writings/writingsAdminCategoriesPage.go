package writings

import (
	"database/sql"
	"errors"
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Categories []*db.WritingCategory
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	categoryRows, err := queries.FetchAllCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("writingCategories Error: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	data.Categories = categoryRows

	CustomWritingsIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "categoriesPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AdminCategoriesModifyPage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	wcid, err := strconv.Atoi(r.PostFormValue("wcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	categoryId, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := queries.UpdateWritingCategory(r.Context(), db.UpdateWritingCategoryParams{
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
		Idwritingcategory: int32(categoryId),
		WritingCategoryID: int32(wcid),
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func AdminCategoriesCreatePage(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	desc := r.PostFormValue("desc")
	pcid, err := strconv.Atoi(r.PostFormValue("pcid"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if err := queries.InsertWritingCategory(r.Context(), db.InsertWritingCategoryParams{
		WritingCategoryID: int32(pcid),
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
	common.TaskDoneAutoRefreshPage(w, r)
}
