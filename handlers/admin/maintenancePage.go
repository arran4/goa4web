package admin

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

type maintenanceTopic struct {
	ID    int32
	Title string
}

// AdminMaintenancePage lists forum topics that only have user grants and can
// be converted to the private handler. The page exposes once off maintenance
// tasks used to migrate old data.
func AdminMaintenancePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Topics   []*maintenanceTopic
		TaskName string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Once Off & Maintenance"
	rows, err := cd.Queries().AdminListTopicsWithUserGrantsNoRoles(r.Context(), true)
	if err != nil {
		log.Printf("list topics with user grants: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	topics := make([]*maintenanceTopic, 0, len(rows))
	for _, row := range rows {
		title := ""
		if row.Title.Valid {
			title = row.Title.String
		}
		topics = append(topics, &maintenanceTopic{ID: row.Idforumtopic, Title: title})
	}
	data := Data{Topics: topics, TaskName: string(TaskForumTopicConvertPrivate)}
	AdminMaintenancePageTmpl.Handle(w, r, data)
}

const AdminMaintenancePageTmpl handlers.Page = "admin/maintenancePage.gohtml"
