package search

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type searchTask struct {
}

const (
	SearchPageTmpl = "searchPage.gohtml"
)

func NewSearchTask() tasks.Task {
	return &searchTask{}
}

func (t *searchTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{tasks.Template(SearchPageTmpl)}
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
