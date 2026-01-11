package imagebbs

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

type imagebbsTask struct {
}

const (
	ImagebbsPageTmpl = "imagebbs/page.gohtml"
)

func NewImagebbsTask() tasks.Task {
	return &imagebbsTask{}
}

func (t *imagebbsTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{ImagebbsPageTmpl}
}

func (t *imagebbsTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *imagebbsTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Board"
	if err := cd.ExecuteSiteTemplate(w, r, ImagebbsPageTmpl, struct{}{}); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
