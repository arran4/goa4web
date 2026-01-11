package search

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

type searchTask struct {
}

const (
	SearchPageTmpl = "searchPage.gohtml"
)

func NewSearchTask() tasks.Task {
	return &searchTask{}
}

func (t *searchTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{SearchPageTmpl}
}

func (t *searchTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *searchTask) Get(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		SearchWords string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Search"
	data := Data{}

	if err := cd.ExecuteSiteTemplate(w, r, SearchPageTmpl, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
