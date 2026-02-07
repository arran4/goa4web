package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/stats"
	"github.com/arran4/goa4web/internal/tasks"
)

const (
	// usageTimeout defines the maximum duration allowed for loading usage statistics
	usageTimeout = 5 * time.Minute
)

type AdminUsageStatsPage struct{}

func (p *AdminUsageStatsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type UserMonthlyUsageDisplay struct {
		*db.UserMonthlyUsageRow
		RowSpan int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Usage Stats"
	queries := cd.Queries()

	ctx, cancel := context.WithTimeout(r.Context(), usageTimeout)
	defer cancel()

	data := stats.BuildUsageStatsData(ctx, queries, cd.CustomQueries(), cd.Config.StatsStartYear)

	AdminUsageStatsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminUsageStatsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Usage Stats", "/admin/usage", &AdminPage{}
}

func (p *AdminUsageStatsPage) PageTitle() string {
	return "Usage Stats"
}

var _ common.Page = (*AdminUsageStatsPage)(nil)
var _ http.Handler = (*AdminUsageStatsPage)(nil)

const AdminUsageStatsPageTmpl tasks.Template = "admin/usageStatsPage.gohtml"
