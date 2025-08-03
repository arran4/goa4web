package linker

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"

	"github.com/gorilla/mux"
)

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Links"
	type Data struct {
		Offset      int32
		HasOffset   bool
		CatId       int32
		CommentOnId int
		ReplyToId   int
	}

	data := Data{}
	if off, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil {
		data.Offset = int32(off)
	}
	data.HasOffset = data.Offset != 0
	if cid, err := strconv.Atoi(r.URL.Query().Get("category")); err == nil {
		data.CatId = int32(cid)
	}
	if cid, err := strconv.Atoi(r.URL.Query().Get("comment")); err == nil {
		data.CommentOnId = cid
	}
	if rid, err := strconv.Atoi(r.URL.Query().Get("reply")); err == nil {
		data.ReplyToId = rid
	}

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
			Name: "User Roles",
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
