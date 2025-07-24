package handlers

import (
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
	data := struct {
		*common.CoreData
		Error   string
		BackURL string
	}{
		CoreData: cd,
		Error:    err.Error(),
		BackURL:  r.Referer(),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "taskErrorAcknowledgementPage.gohtml", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
