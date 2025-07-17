package common

import (
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
)

func TaskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*CoreData),
	}
	data.AutoRefresh = true

	TemplateHandler(w, r, "taskDoneAutoRefreshPage.gohtml", data)
}
