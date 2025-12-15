package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
)

func Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		return
	}
	cd.PageTitle = "Private Forum"
	forumhandlers.CreateTopicPageWithPostTask(w, r, TaskPrivateTopicCreate, &forumhandlers.CreateTopicPageForm{})
}
