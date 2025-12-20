package privateforum

import (
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"
)

type privateForumTask struct {
}

const (
	CreateTopicTmpl = "forum/create_topic.gohtml"
	TopicsOnlyTmpl  = "privateforum/topics_only.gohtml"
)

func NewPrivateForumTask() tasks.Task {
	return &privateForumTask{}
}

func (t *privateForumTask) TemplatesRequired() []string {
	return []string{CreateTopicTmpl, TopicsOnlyTmpl}
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
	// Show topics only on the main private page (no creation form)
	handlers.TemplateHandler(w, r, TopicsOnlyTmpl, nil)
}
