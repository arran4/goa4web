package linker

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
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
		Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRowRow
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	data.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	data.HasOffset = data.Offset != 0
	vars := mux.Vars(r)
	data.CatId, _ = strconv.Atoi(vars["category"])
	data.CommentOnId, _ = strconv.Atoi(r.URL.Query().Get("comment"))
	data.ReplyToId, _ = strconv.Atoi(r.URL.Query().Get("reply"))

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)

	uid := data.CoreData.UserID
	linkerPosts, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow(r.Context(), db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRowParams{
		ViewerID:         uid,
		Idlinkercategory: int32(data.CatId),
		ViewerUserID:     sql.NullInt32{Int32: uid, Valid: uid != 0},
		Limit:            15,
		Offset:           int32(data.Offset),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	for _, row := range linkerPosts {
		if !data.CoreData.HasGrant("linker", "link", "see", row.Idlinker) {
			continue
		}
		data.Links = append(data.Links, row)
	}

	handlers.TemplateHandler(w, r, "linkerCategoryPage", data)
}
