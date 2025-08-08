package privateforum

import (
	"log"
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
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		if err := cd.ExecuteSiteTemplate(w, r, "noAccessPage.gohtml", struct{}{}); err != nil {
			log.Printf("render no access page: %v", err)
		}
		return
	}
	cd.PageTitle = "Private Forum"
	data := Data{CreateTask: TaskPrivateTopicCreate}
	handlers.TemplateHandler(w, r, "privateForumPage", data)
}
