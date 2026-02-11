package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/stats"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminServerStatsPage struct {
	Srv        *server.Server
	ConfigFile string
}

func (p *AdminServerStatsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Server Stats"
	var dlqReg *dlq.Registry
	var emailReg *email.Registry
	var routerReg *router.Registry
	if p.Srv != nil {
		dlqReg = p.Srv.DLQReg
		emailReg = p.Srv.EmailReg
		routerReg = p.Srv.RouterReg
	}
	var routerModules []string
	if routerReg != nil {
		routerModules = routerReg.Names()
	}
	data := stats.BuildServerStatsData(cd.Config, p.ConfigFile, cd.TasksReg, cd.DBRegistry(), dlqReg, emailReg, routerModules)

	AdminServerStatsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminServerStatsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Server Stats", "/admin/stats", &AdminPage{}
}

func (p *AdminServerStatsPage) PageTitle() string {
	return "Server Stats"
}

var _ common.Page = (*AdminServerStatsPage)(nil)
var _ http.Handler = (*AdminServerStatsPage)(nil)

const AdminServerStatsPageTmpl tasks.Template = "admin/serverStatsPage.gohtml"
