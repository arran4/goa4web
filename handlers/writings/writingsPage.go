package writings

import (
	"github.com/arran4/goa4web/core/common"
	"net/http"
)

func WritingsPage(w http.ResponseWriter, r *http.Request) {
	t := NewWritingsTask().(*writingsTask)
	t.Get(w, r)
}
func CustomWritingsIndex(data *common.CoreData, r *http.Request) {
	data.CustomIndexItems = append(data.CustomIndexItems, WritingsGeneralIndexItems(data, r)...)
}
