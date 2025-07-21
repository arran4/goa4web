package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TaskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	data.AutoRefresh = true

	TemplateHandler(w, r, "taskDoneAutoRefreshPage.gohtml", data)
}
