package privateforum

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

func Page(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CreateTask tasks.TaskString
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Private Forum"
	data := Data{CreateTask: TaskPrivateTopicCreate}
	handlers.TemplateHandler(w, r, "privateForumPage", data)
}
