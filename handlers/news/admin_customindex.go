package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
)

// CustomAdminNewsIndex injects pagination links for the admin news pages.
func CustomAdminNewsIndex(cd *common.CoreData, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	if offset != 0 {
		cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
			Name: "The start",
			Link: "/admin/news?offset=0",
		})
	}
	cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
		Name: fmt.Sprintf("Next %d", ps),
		Link: fmt.Sprintf("/admin/news?offset=%d", offset+ps),
	})
	if offset > 0 {
		cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
			Name: fmt.Sprintf("Previous %d", ps),
			Link: fmt.Sprintf("/admin/news?offset=%d", offset-ps),
		})
	}
}
