package common

import (
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func TemplatePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(KeyCoreData).(*CoreData),
	}

	if err := templates.RenderTemplate(w, "templatePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
