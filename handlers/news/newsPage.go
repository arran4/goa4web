package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
)

func CustomNewsIndex(data *common.CoreData, r *http.Request) {
	data.RSSFeedUrl = "/news.rss"
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "RSS Feed",
		Link: "/news.rss",
	})
	userHasAdmin := data.HasGrant("news", "post", "edit", 0) && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "User Permissions",
			Link: "/admin/news/user/permissions",
		})
	}
	userHasWriter := data.HasGrant("news", "post", "post", 0)
	if userHasWriter || userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Add News",
			Link: "/news/post",
		})
	}

	vars := mux.Vars(r)
	newsId := vars["news"]
	if newsId != "" {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Return to list",
			Link: fmt.Sprintf("/?offset=%d", 0),
		})
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset != 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "The start",
			Link: fmt.Sprintf("?offset=%d", 0),
		})
	}
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Next 10",
		Link: fmt.Sprintf("?offset=%d", offset+10),
	})
	if offset > 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "Previous 10",
			Link: fmt.Sprintf("?offset=%d", offset-10),
		})
	}
}
