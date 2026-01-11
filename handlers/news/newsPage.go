package news

import (
	"github.com/arran4/goa4web/core/common"
	"net/http"
)

func NewsPageHandler(w http.ResponseWriter, r *http.Request) {
	t := NewNewsTask().(*newsTask)
	t.Get(w, r)
}

func CustomNewsIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = append(data.CustomIndexItems, NewsGeneralIndexItems(data, r)...)
}
