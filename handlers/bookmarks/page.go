package bookmarks

import (
	"github.com/arran4/goa4web/core/common"
	"net/http"
)

func BookmarksPage(w http.ResponseWriter, r *http.Request) {
	t := NewBookmarksTask().(*bookmarksTask)
	t.Get(w, r)
}

func bookmarksCustomIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = []common.IndexItem{}
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Show",
		Link: "/bookmarks/mine",
	})
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: "Edit",
		Link: "/bookmarks/edit",
	})
}
