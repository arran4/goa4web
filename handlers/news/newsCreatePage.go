package news

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

const (
	NewsCreatePageTmpl handlers.Page = "news/createPage.gohtml"
)

func NewsCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Add News"

	// Permission check
	if !CanPostNews(cd) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	NewsCreatePageTmpl.Handle(w, r, nil)
}
