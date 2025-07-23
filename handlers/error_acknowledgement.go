package handlers

import (
	"net/http"
)

// TaskErrorAcknowledgementPage renders a page displaying an error message.
func TaskErrorAcknowledgementPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Error   string
		BackURL string
	}
	data := Data{
		Error:   r.URL.Query().Get("error"),
		BackURL: r.Referer(),
	}
	if data.Error == "" {
		data.Error = r.PostFormValue("error")
	}
	TemplateHandler(w, r, "taskErrorAcknowledgementPage.gohtml", data)
}
