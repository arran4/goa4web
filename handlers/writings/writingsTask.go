package writings

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/tasks"
)

type writingsTask struct {
}

const (
	WritingsPageTmpl handlers.Page = "writings/page.gohtml"
)

func NewWritingsTask() tasks.Task {
	return &writingsTask{}
}

func (t *writingsTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{WritingsPageTmpl}
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
		Image:       share.MakeImageURL(cd.AbsoluteURL(), "Writings", cd.ShareSigner, false),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "website",
	}

	WritingsPageTmpl.Handle(w, r, data)
}
