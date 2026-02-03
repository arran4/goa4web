package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/stats"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

const (
	// usageTimeout defines the maximum duration allowed for loading usage statistics
	usageTimeout = 5 * time.Minute
)

func AdminUsageStatsPage(w http.ResponseWriter, r *http.Request) {
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

	AdminUsageStatsPageTmpl.Handle(w, r, data)
}

const AdminUsageStatsPageTmpl tasks.Template = "admin/usageStatsPage.gohtml"
