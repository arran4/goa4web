package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// TaskErrorAcknowledgementPage renders a page displaying an error message.
func TaskErrorAcknowledgementPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Error   string
		BackURL string
	}
	cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd != nil {
		cd.PageTitle = "Error"
	}
	data := Data{
		Error:   r.URL.Query().Get("error"),
		BackURL: r.Referer(),
	}
	if data.Error == "" {
		data.Error = r.PostFormValue("error")
	}
	TaskErrorAcknowledgementPageTmpl.Handle(w, r, data)
}
