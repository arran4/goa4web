package news

import (
	"fmt"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strconv"
)

type newsTask struct {
}

const (
	NewsPageTmpl = "news/page.gohtml"
)

func NewNewsTask() tasks.Task {
	return &newsTask{}
}

func (t *newsTask) TemplatesRequired() []string {
	return []string{NewsPageTmpl}
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
		Image:       cd.AbsoluteURL(fmt.Sprintf("/api/og-image?title=%s", "News")),
		URL:         cd.AbsoluteURL(r.URL.String()),
	}

	if err := cd.ExecuteSiteTemplate(w, r, NewsPageTmpl, struct{}{}); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
