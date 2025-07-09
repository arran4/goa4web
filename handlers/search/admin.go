package search

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func adminSearchPage(w http.ResponseWriter, r *http.Request) {
	type Stats struct {
		Words    int64
		WordList int64
		Comments int64
		News     int64
		Blogs    int64
		Linker   int64
		Writing  int64
		Writings int64
		Images   int64
	}

	type Data struct {
		*hcommon.CoreData
		Stats Stats
	}

	data := Data{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	ctx := r.Context()
	count := func(query string, dest *int64) {
		if err := queries.DB().QueryRowContext(ctx, query).Scan(dest); err != nil && err != sql.ErrNoRows {
			log.Printf("adminSearchPage count query error: %v", err)
		}
	}

	count("SELECT COUNT(*) FROM searchwordlist", &data.Stats.Words)
	count("SELECT COUNT(*) FROM comments_search", &data.Stats.Comments)
	count("SELECT COUNT(*) FROM site_news_search", &data.Stats.News)
	count("SELECT COUNT(*) FROM blogs_search", &data.Stats.Blogs)
	count("SELECT COUNT(*) FROM linker_search", &data.Stats.Linker)
	count("SELECT COUNT(*) FROM writing_search", &data.Stats.Writing)
	count("SELECT COUNT(*) FROM writing_search", &data.Stats.Writings)
	count("SELECT COUNT(*) FROM imagepost_search", &data.Stats.Images)

	if err := templates.RenderTemplate(w, "adminSearchPage", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminSearchRemakeCommentsSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := struct {
		*hcommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteCommentsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteCommentsSearch: %w", err).Error())
	}
	if err := queries.RemakeCommentsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeCommentsSearchInsert: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func adminSearchRemakeNewsSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := struct {
		*hcommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteSiteNewsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteSiteNewsSearch: %w", err).Error())
	}
	if err := queries.RemakeNewsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeNewsSearchInsert: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func adminSearchRemakeBlogSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := struct {
		*hcommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteBlogsSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteBlogsSearch: %w", err).Error())
	}
	if err := queries.RemakeBlogsSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeBlogsSearchInsert: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func adminSearchRemakeLinkerSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := struct {
		*hcommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteLinkerSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteLinkerSearch: %w", err).Error())
	}
	if err := queries.RemakeLinkerSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeLinkerSearchInsert: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func adminSearchRemakeWritingSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := struct {
		*hcommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteWritingSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteWritingSearch: %w", err).Error())
	}
	if err := queries.RemakeWritingSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeWritingSearchInsert: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminSearchRemakeImageSearchPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	data := struct {
		*hcommon.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData),
		Back:     "/admin/search",
	}
	if err := queries.DeleteImagePostSearch(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("DeleteImagePostSearch: %w", err).Error())
	}
	if err := queries.RemakeImagePostSearchInsert(r.Context()); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("RemakeImagePostSearchInsert: %w", err).Error())
	}

	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
