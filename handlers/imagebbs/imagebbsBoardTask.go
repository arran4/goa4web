package imagebbs

import (
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type imagebbsBoardTask struct {
}

var _ tasks.Task = (*imagebbsBoardTask)(nil)

const (
	ImagebbsBoardPageTmpl = "imagebbs/boardPage.gohtml"
)

func NewImagebbsBoardTask() tasks.Task {
	return &imagebbsBoardTask{}
}

func (t *imagebbsBoardTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{tasks.Template(ImagebbsBoardPageTmpl)}
}

func (t *imagebbsBoardTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *imagebbsBoardTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	bid := cd.SelectedBoardID()

	if !cd.HasGrant("imagebbs", "board", "view", bid) {
		fmt.Println("TODO: FIx: Add enforced Access in router rather than task")
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}

	cd.PageTitle = fmt.Sprintf("Board %d", bid)
	if err := cd.ExecuteSiteTemplate(w, r, ImagebbsBoardPageTmpl, struct{}{}); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
