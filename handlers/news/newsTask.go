package news

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/tasks"
)

type newsTask struct {
}

const (
	NewsPageTmpl tasks.Template = "news/page.gohtml"
)

func NewNewsTask() tasks.Task {
	return &newsTask{}
}

func (t *newsTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{NewsPageTmpl}
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

	img, err := share.MakeImageURL(cd.AbsoluteURL(), "Latest News", "Latest news and announcements.", cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making image URL: %v", err)
	}

	cd.OpenGraph = &common.OpenGraph{
		Title:       "News",
		Description: "Latest news and announcements.",
		Image:       img,
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
		URL:         cd.AbsoluteURL(r.URL.String()),
		Type:        "website",
	}

	NewsPageTmpl.Handle(w, r, struct{}{})
}
