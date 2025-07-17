package linker

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func CategoryPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Offset      int
		HasOffset   bool
		CatId       int
		CommentOnId int
		ReplyToId   int
		Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	data.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	data.HasOffset = data.Offset != 0
	vars := mux.Vars(r)
	data.CatId, _ = strconv.Atoi(vars["category"])
	data.CommentOnId, _ = strconv.Atoi(r.URL.Query().Get("comment"))
	data.ReplyToId, _ = strconv.Atoi(r.URL.Query().Get("reply"))

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	linkerPosts, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(r.Context(), db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{Idlinkercategory: int32(data.CatId)})
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

	handlers.TemplateHandler(w, r, "linkerCategoryPage", data)
}
