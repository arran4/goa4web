package handlers

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
)

func TaskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}
	data := Data{
		CoreData: r.Context().Value(common.ContextValues("coreData")).(*common.CoreData),
	}
	data.AutoRefresh = true

	TemplateHandler(w, r, "taskDoneAutoRefreshPage.gohtml", data)
}
