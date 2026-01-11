package news

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/tasks"
)

type newsTask struct {
}

const (
	NewsPageTmpl handlers.Page = "news/page.gohtml"
)

func NewNewsTask() tasks.Task {
	return &newsTask{}
}

func (t *newsTask) TemplatesRequired() []tasks.Page {
	return []tasks.Page{NewsPageTmpl}
}

func (t *newsTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *newsTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("?offset=%d", offset-ps)
		cd.StartLink = "?offset=0"
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       "News",
		Description: "Latest news and announcements.",
		Image:       share.MakeImageURL(cd.AbsoluteURL(""), "Latest News", cd.ShareSigner, false),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
	}

	NewsPageTmpl.Handle(w, r, struct{}{})
}
