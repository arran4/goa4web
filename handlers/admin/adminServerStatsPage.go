package admin

import (
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/router"
)

func (h *Handlers) AdminServerStatsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Server Stats"
	var dlqReg, emailReg, routerReg = (*dlq.Registry)(nil), (*email.Registry)(nil), (*router.Registry)(nil)
	if h.Srv != nil {
		dlqReg = h.Srv.DLQReg
		emailReg = h.Srv.EmailReg
		routerReg = h.Srv.RouterReg
	}
	data := BuildServerStatsData(cd.Config, h.ConfigFile, cd.TasksReg, cd.DBRegistry(), dlqReg, emailReg, routerReg)

	AdminServerStatsPageTmpl.Handle(w, r, data)
}

const AdminServerStatsPageTmpl tasks.Template = "admin/serverStatsPage.gohtml"
