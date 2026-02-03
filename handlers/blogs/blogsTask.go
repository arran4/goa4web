package blogs

import (
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

type blogsTask struct {
}

var _ tasks.Task = (*blogsTask)(nil)

const (
	BlogsPageTmpl tasks.Template = "blogs/page.gohtml"
)

func NewBlogsTask() tasks.Task {
	return &blogsTask{}
}

func (t *blogsTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{BlogsPageTmpl}
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

	BlogsPageTmpl.Handle(w, r, struct{}{})
}
