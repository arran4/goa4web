package admin

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func AdminPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Users        int64
		Languages    int64
		News         int64
		Blogs        int64
		ForumTopics  int64
		ForumThreads int64
		Writings     int64
	}

	type Data struct {
		*common.CoreData
		AdminSections []common.AdminSection
		Stats         Stats
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Admin"
	data := Data{
		CoreData:      cd,
		AdminSections: cd.Nav.AdminSections(),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	stats, err := queries.AdminGetDashboardStats(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	data.Stats.Users = stats.Users
	data.Stats.Languages = stats.Languages
	data.Stats.News = stats.News
	data.Stats.Blogs = stats.Blogs
	data.Stats.ForumTopics = stats.ForumTopics
	data.Stats.ForumThreads = stats.ForumThreads
	data.Stats.Writings = stats.Writings

	handlers.TemplateHandler(w, r, "adminPage", data)
}
