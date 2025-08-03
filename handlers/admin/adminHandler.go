package admin

import (
	_ "embed"
	"github.com/arran4/goa4web/core/consts"
	"net/http"

	"github.com/arran4/goa4web/core/common"
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
	counts, err := queries.AdminSiteCounts(r.Context())
	if err != nil {
		http.Error(w, "database not available", http.StatusInternalServerError)
		return
	}
	data.Stats.Users = counts.Users
	data.Stats.Languages = counts.Languages
	data.Stats.News = counts.News
	data.Stats.Blogs = counts.Blogs
	data.Stats.ForumTopics = counts.ForumTopics
	data.Stats.ForumThreads = counts.ForumThreads
	data.Stats.Writings = counts.Writings

	handlers.TemplateHandler(w, r, "adminPage", data)
}
