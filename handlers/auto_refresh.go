package handlers

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
)

func TaskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct{}
	data := Data{}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Done"
	cd.AutoRefresh = "1"
	TemplateHandler(w, r, "taskDoneAutoRefreshPage.gohtml", data)
}
