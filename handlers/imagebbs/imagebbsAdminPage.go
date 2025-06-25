package imagebbs

import (
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
	}

	CustomImageBBSIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
