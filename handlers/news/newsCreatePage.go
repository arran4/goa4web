package news

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

const (
	NewsCreatePageTmpl = "news/createPage.gohtml"
)

func NewsCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Add News"

	// Permission check
	if !common.Allowed(r, "content writer", "administrator") {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	if err := cd.ExecuteSiteTemplate(w, r, NewsCreatePageTmpl, nil); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
