package admin

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

type maintenanceTopic struct {
	ID    int32
	Title string
}

type AdminMaintenancePage struct{}

func (p *AdminMaintenancePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Topics   []*maintenanceTopic
		TaskName string
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Once Off & Maintenance"
	rows, err := cd.Queries().AdminListTopicsWithUserGrantsNoRoles(r.Context(), true)
	if err != nil {
		log.Printf("list topics with user grants: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
	AdminMaintenancePageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminMaintenancePage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Maintenance", "/admin/maintenance", &AdminPage{}
}

func (p *AdminMaintenancePage) PageTitle() string {
	return "Once Off & Maintenance"
}

var _ common.Page = (*AdminMaintenancePage)(nil)
var _ http.Handler = (*AdminMaintenancePage)(nil)

const AdminMaintenancePageTmpl tasks.Template = "admin/maintenancePage.gohtml"
