package handlers

import "net/http"

func TaskDoneAutoRefreshPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}
	data := Data{
		CoreData: r.Context().Value(ContextKey("coreData")).(*CoreData),
	}
	data.AutoRefresh = true

	TemplateHandler(w, r, "taskDoneAutoRefreshPage.gohtml", data)
}
