package templates

import (
	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

// Page renders a simple template example.
func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData),
	}

	if err := templates.RenderTemplate(w, "templatePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
