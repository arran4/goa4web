package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func NewsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("?offset=%d", offset-ps)
		cd.StartLink = "?offset=0"
	}
	handlers.TemplateHandler(w, r, "newsPage", struct{}{})
}

func CustomNewsIndex(data *common.CoreData, r *http.Request) {
	data.RSSFeedURL = "/news.rss"
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "RSS Feed",
		Link: "/news.rss",
	})
	userHasAdmin := data.HasGrant("news", "post", "edit", 0) && data.AdminMode
	if userHasAdmin {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: "User Roles",
			Link: "/admin/news/users/roles",
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
}
