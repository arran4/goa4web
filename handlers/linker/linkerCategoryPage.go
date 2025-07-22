package linker

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

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

	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		data.Offset = off
	}
	data.HasOffset = data.Offset != 0
	vars := mux.Vars(r)
	if cid, err := strconv.Atoi(vars["category"]); err == nil {
		data.CatId = cid
	}
	if cid, err := strconv.Atoi(r.URL.Query().Get("comment")); err == nil {
		data.CommentOnId = cid
	}
	if rid, err := strconv.Atoi(r.URL.Query().Get("reply")); err == nil {
		data.ReplyToId = rid
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

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
