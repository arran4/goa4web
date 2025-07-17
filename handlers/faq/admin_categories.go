package faq

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func AdminCategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecorecommon.CoreData
		Rows []*db.GetFAQCategoriesWithQuestionCountRow
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecorecommon.CoreData),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	rows, err := queries.GetFAQCategoriesWithQuestionCount(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	common.TemplateHandler(w, r, "adminCategoriesPage.gohtml", data)
}

func CategoriesRenameActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	if err := queries.RenameFAQCategory(r.Context(), db.RenameFAQCategoryParams{
		Name: sql.NullString{
			String: text,
			Valid:  true,
		},
		Idfaqcategories: int32(cid),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func CategoriesDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(r.PostFormValue("cid"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	if err := queries.DeleteFAQCategory(r.Context(), int32(cid)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func CategoriesCreateActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("cname")
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	if err := queries.CreateFAQCategory(r.Context(), sql.NullString{
		String: text,
		Valid:  true,
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}
