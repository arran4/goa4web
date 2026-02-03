package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/stats"
	"github.com/arran4/goa4web/internal/tasks"
)

func (h *Handlers) AdminServerStatsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Server Stats"
	var dlqReg *dlq.Registry
	var emailReg *email.Registry
	var routerReg *router.Registry
	if h.Srv != nil {
		dlqReg = h.Srv.DLQReg
		emailReg = h.Srv.EmailReg
		routerReg = h.Srv.RouterReg
	}
	var routerModules []string
	if routerReg != nil {
		routerModules = routerReg.Names()
	}
	data := stats.BuildServerStatsData(cd.Config, h.ConfigFile, cd.TasksReg, cd.DBRegistry(), dlqReg, emailReg, routerModules)

	AdminServerStatsPageTmpl.Handle(w, r, data)
}

const AdminServerStatsPageTmpl tasks.Template = "admin/serverStatsPage.gohtml"
