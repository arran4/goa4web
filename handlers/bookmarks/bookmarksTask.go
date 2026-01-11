package bookmarks

import (
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

type bookmarksTask struct {
}

const (
	BookmarksPageTmpl = "bookmarks/page.gohtml"
	InfoPageTmpl      = "bookmarks/infoPage.gohtml"
)

func NewBookmarksTask() tasks.Task {
	return &bookmarksTask{}
}

func (t *bookmarksTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{BookmarksPageTmpl, InfoPageTmpl}
}

func (t *bookmarksTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *bookmarksTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	cd.PageTitle = "Bookmarks"

	if uid == 0 {
		if err := cd.ExecuteSiteTemplate(w, r, InfoPageTmpl, struct{}{}); err != nil {
			handlers.RenderErrorPage(w, r, err)
		}
		return
	}

	if err := cd.ExecuteSiteTemplate(w, r, BookmarksPageTmpl, struct{}{}); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
