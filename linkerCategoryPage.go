package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func linkerCategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Offset      int
		CatId       int
		CommentOnId int
		ReplyToId   int
		Links       []*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	data.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	vars := mux.Vars(r)
	data.CatId, _ = strconv.Atoi(vars["category"])
	data.CommentOnId, _ = strconv.Atoi(r.URL.Query().Get("comment"))
	data.ReplyToId, _ = strconv.Atoi(r.URL.Query().Get("reply"))

	queries := r.Context().Value(common.KeyQueries).(*Queries)

	linkerPosts, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{Idlinkercategory: int32(data.CatId)})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Links = linkerPosts

	CustomLinkerIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "categoryPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
