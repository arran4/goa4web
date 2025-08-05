package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// RenderErrorPage displays err using the standard error acknowledgment page.
func RenderErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	cd, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil {
		cd = &common.CoreData{}
	}
	cd.PageTitle = "Error"
	data := struct {
		*common.CoreData
		Error   string
		BackURL string
	}{
		CoreData: cd,
		Error:    err.Error(),
		BackURL:  r.Referer(),
	}
	contentType := w.Header().Get("Content-Type")

	status := http.StatusInternalServerError
	var he *HTTPError
	if errors.As(err, &he) {
		status = he.Status
	} else {
		switch err.Error() {
		case "Forbidden":
			status = http.StatusForbidden
		case "Unauthorized":
			status = http.StatusUnauthorized
		case "Bad Request":
			status = http.StatusBadRequest
		case "Not Found":
			status = http.StatusNotFound
		}
	}
	w.WriteHeader(status)

	if err := cd.ExecuteSiteTemplate(w, r, "taskErrorAcknowledgementPage.gohtml", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
}
