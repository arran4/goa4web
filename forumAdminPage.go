package goa4web

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/templates"
)

func forumAdminPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Categories int64
		Topics     int64
		Threads    int64
	}

	type Data struct {
		*CoreData
		Stats Stats
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	ctx := r.Context()
	count := func(q string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, q).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("forumAdminPage count query error: %v", err)
		}
	}

	count("SELECT COUNT(*) FROM forumcategory", &data.Stats.Categories)
	count("SELECT COUNT(*) FROM forumtopic", &data.Stats.Topics)
	count("SELECT COUNT(*) FROM forumthread", &data.Stats.Threads)

	CustomForumIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "adminPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
