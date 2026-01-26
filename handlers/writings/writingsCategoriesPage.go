package writings

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func CategoriesPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Request           *http.Request
		CategoryId        int32
		WritingCategoryID int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Writing Categories"
	data := Data{Request: r}
	WritingsCategoriesPageTmpl.Handle(w, r, data)
}

const WritingsCategoriesPageTmpl tasks.Template = "writings/writingsCategoriesPage.gohtml"
