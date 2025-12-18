package news

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/gorilla/mux"
	"net/http"
)

func NewsPageHandler(w http.ResponseWriter, r *http.Request) {
	t := NewNewsTask().(*newsTask)
	t.Get(w, r)
}

func CustomNewsIndex(data *common.CoreData, r *http.Request) {
	data.RSSFeedURL = "/news.rss"
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name:   "RSS Feed",
		Link:   "/news.rss",
		Folded: true,
	})
	userHasWriter := data.HasGrant("news", "post", "post", 0)
	if userHasWriter {
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
