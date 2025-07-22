package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"

	"github.com/gorilla/mux"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Offset      int
		HasOffset   bool
		CatId       int
		CommentOnId int
		ReplyToId   int
		Links       []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow
		Categories  []*db.LinkerCategory
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		data.Offset = off
	}
	data.HasOffset = data.Offset != 0
	if cid, err := strconv.Atoi(r.URL.Query().Get("category")); err == nil {
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
	linkerPosts, err := queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated(r.Context(), db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedParams{
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

	categories, err := queries.GetAllLinkerCategoriesForUser(r.Context(), db.GetAllLinkerCategoriesForUserParams{
		ViewerID:     uid,
		ViewerUserID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerCategories Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Categories = categories

	handlers.TemplateHandler(w, r, "linkerPage", data)
}

func CustomLinkerIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	if r.URL.Path == "/linker" || strings.HasPrefix(r.URL.Path, "/linker/category/") {
		data.RSSFeedUrl = "/linker/rss"
		data.AtomFeedUrl = "/linker/atom"
	}

	userHasAdmin := data.HasRole("administrator") && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "User Permissions",
			Link: "/admin/linker/users/roles",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Category Controls",
			Link: "/admin/linker/categories",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Approve links",
			Link: "/admin/linker/queue",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Add link",
			Link: "/admin/linker/add",
		})
	}
	vars := mux.Vars(r)
	categoryId := vars["category"]
	offset := 0
	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		offset = off
	}
	if categoryId == "" {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/linker?offset=%d", offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/linker?offset=%d", offset-15),
			})
		}
	} else {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/linker/category/%s?offset=%d", categoryId, offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/linker/category/%s?offset=%d", categoryId, offset-15),
			})
		}
	}

}
