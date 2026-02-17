package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/tasks"
)

// StartGroupDiscussionPage renders a dedicated page to start a private group discussion.
func StartGroupDiscussionPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	// TODO: FIx: Add enforced Access in router rather than task
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	// Page title/header as requested
	cd.PageTitle = "Start private group discussion"

	// Prepare data matching forum.CreateTopicPageWithPostTask
	data := struct {
		CreateTask tasks.TaskString
		FormData   *forumhandlers.CreateTopicPageForm
	}{CreateTask: TaskPrivateTopicCreate, FormData: &forumhandlers.CreateTopicPageForm{}}
	PrivateForumStartDiscussionPageTmpl.Handle(w, r, data)
}

const PrivateForumStartDiscussionPageTmpl tasks.Template = "privateforum/start_discussion.gohtml"
