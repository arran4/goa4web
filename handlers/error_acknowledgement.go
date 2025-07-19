package handlers

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
)

// TaskErrorAcknowledgementPage renders a page displaying an error message.
func TaskErrorAcknowledgementPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Error   string
		BackURL string
	}
	data := Data{
		CoreData: r.Context().Value(common.ContextValues("coreData")).(*common.CoreData),
		Error:    r.URL.Query().Get("error"),
		BackURL:  r.Referer(),
	}
	if data.Error == "" {
		data.Error = r.PostFormValue("error")
	}
	TemplateHandler(w, r, "tasks/error_ack.gohtml", data)
}
