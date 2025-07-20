package handlers

import (
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// TaskErrorAcknowledgementPage renders a page displaying an error message.
func TaskErrorAcknowledgementPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Error   string
		BackURL string
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Error:    r.URL.Query().Get("error"),
		BackURL:  r.Referer(),
	}
	if data.Error == "" {
		data.Error = r.PostFormValue("error")
	}
	TemplateHandler(w, r, "taskErrorAcknowledgementPage.gohtml", data)
}
