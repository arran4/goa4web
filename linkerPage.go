package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
	"github.com/gorilla/mux"
)

func linkerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Offset      int
		CatId       int
		CommentOnId int
		ReplyToId   int
		Links       []*GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow
		Categories  []*Linkercategory
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
	}

	data.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	data.CatId, _ = strconv.Atoi(r.URL.Query().Get("category"))
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

	categories, err := queries.GetAllLinkerCategories(r.Context())
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

	CustomLinkerIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "page.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func CustomLinkerIndex(data *CoreData, r *http.Request) {
	if r.URL.Path == "/linker" || strings.HasPrefix(r.URL.Path, "/linker/category/") {
		data.RSSFeedUrl = "/linker/rss"
		data.AtomFeedUrl = "/linker/atom"
	}

	userHasAdmin := data.HasRole("administrator")
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "User Permissions",
			Link: "/admin/linker/users/levels",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Category Controls",
			Link: "/admin/linker/categories",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Approve links",
			Link: "/admin/linker/queue",
		})
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Add link",
			Link: "/admin/linker/add",
		})
	}
	vars := mux.Vars(r)
	categoryId := vars["category"]
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if categoryId == "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/linker?offset=%d", offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/linker?offset=%d", offset-15),
			})
		}
	} else {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/linker/category/%s?offset=%d", categoryId, offset+15),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/linker/category/%s?offset=%d", categoryId, offset-15),
			})
		}
	}

}
