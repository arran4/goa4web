package common

import (
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

// TaskErrorAcknowledgementPage renders a page displaying an error message.
func TaskErrorAcknowledgementPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Error   string
		BackURL string
	}
	data := Data{
		CoreData: r.Context().Value(ContextKey("coreData")).(*CoreData),
		Error:    r.URL.Query().Get("error"),
		BackURL:  r.Referer(),
	}
	if data.Error == "" {
		data.Error = r.PostFormValue("error")
	}
	if err := templates.RenderTemplate(w, "taskErrorAcknowledgementPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
