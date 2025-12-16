package privateforum

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

type privateForumTask struct {
}

const (
	CreateTopicTmpl = "forum/create_topic.gohtml"
)

func NewPrivateForumTask() tasks.Task {
	return &privateForumTask{}
}

func (t *privateForumTask) TemplatesRequired() []string {
	return []string{CreateTopicTmpl}
}

func (t *privateForumTask) Action(w http.ResponseWriter, r *http.Request) any {
	return nil
}

func (t *privateForumTask) Get(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	cd.PageTitle = "Private Forum"
	forumhandlers.CreateTopicPageWithPostTask(w, r, TaskPrivateTopicCreate, &forumhandlers.CreateTopicPageForm{})
}
