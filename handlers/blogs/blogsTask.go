package blogs

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
	"strconv"
)

type blogsTask struct {
}

const (
	BlogsPageTmpl = "blogs/page.gohtml"
)

func NewBlogsTask() tasks.Task {
	return &blogsTask{}
}

func (t *blogsTask) TemplatesRequired() []string {
	return []string{BlogsPageTmpl}
}

func (t *blogsTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *blogsTask) Get(w http.ResponseWriter, r *http.Request) {
	buid := r.URL.Query().Get("uid")
	userID, _ := strconv.Atoi(buid)

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Blogs"
	cd.SetCurrentProfileUserID(int32(userID))

	offset := cd.Offset()
	ps := cd.PageSize()
	qv := r.URL.Query()
	qv.Set("offset", strconv.Itoa(offset+ps))
	cd.NextLink = "/blogs?" + qv.Encode()
	if offset > 0 {
		qv.Set("offset", strconv.Itoa(offset-ps))
		cd.PrevLink = "/blogs?" + qv.Encode()
	}

	if err := cd.ExecuteSiteTemplate(w, r, BlogsPageTmpl, struct{}{}); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
