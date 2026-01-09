package writings

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strconv"
)

type writingsTask struct {
}

const (
	WritingsPageTmpl = "writings/page.gohtml"
)

func NewWritingsTask() tasks.Task {
	return &writingsTask{}
}

func (t *writingsTask) TemplatesRequired() []string {
	return []string{WritingsPageTmpl}
}

func (t *writingsTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *writingsTask) Get(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CategoryId        int32
		WritingCategoryID int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Writings"
	data := Data{}
	data.CategoryId = 0
	data.WritingCategoryID = data.CategoryId

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	qv := r.URL.Query()
	qv.Set("offset", strconv.Itoa(offset+ps))
	cd.NextLink = "/writings?" + qv.Encode()
	if offset > 0 {
		qv.Set("offset", strconv.Itoa(offset-ps))
		cd.PrevLink = "/writings?" + qv.Encode()
		cd.StartLink = "/writings?offset=0"
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       "Writings",
		Description: "A collection of articles and long-form content.",
		Image:       cd.AbsoluteURL(fmt.Sprintf("/api/og-image?title=%s", "Writings")),
		URL:         cd.AbsoluteURL(r.URL.String()),
	}

	if err := cd.ExecuteSiteTemplate(w, r, WritingsPageTmpl, data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
