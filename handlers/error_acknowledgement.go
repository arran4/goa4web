package handlers

import "net/http"

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
	TemplateHandler(w, r, "taskErrorAcknowledgementPage.gohtml", data)
}
