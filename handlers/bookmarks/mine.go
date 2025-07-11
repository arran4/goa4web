package bookmarks

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

type BookmarkEntry struct {
	Url  string
	Name string
}

type BookmarkCategory struct {
	Name    string
	Entries []*BookmarkEntry
}

type BookmarkColumn struct {
	Categories []*BookmarkCategory
}

func preprocessBookmarks(bookmarks string) []*BookmarkColumn {
	lines := strings.Split(bookmarks, "\n")
	var result = []*BookmarkColumn{{}}
	var currentCategory *BookmarkCategory

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.EqualFold(line, "column") {
			result = append(result, &BookmarkColumn{})
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		if len(parts) > 0 && strings.EqualFold(parts[0], "Category:") {
			categoryName := strings.Join(parts[1:], " ")
			if currentCategory == nil {
				currentCategory = &BookmarkCategory{Name: categoryName}
			} else if currentCategory.Name != "" {
				result[len(result)-1].Categories = append(result[len(result)-1].Categories, currentCategory)
				currentCategory = &BookmarkCategory{Name: categoryName}
			} else {
				currentCategory.Name = categoryName
			}
		} else if len(parts) > 0 && currentCategory != nil {
			var entry BookmarkEntry
			entry.Url = parts[0]
			entry.Name = parts[0]
			if len(parts) > 1 {
				entry.Name = strings.Join(parts[1:], " ")
			}
			currentCategory.Entries = append(currentCategory.Entries, &entry)
		}
	}

	if currentCategory != nil && currentCategory.Name != "" {
		result[len(result)-1].Categories = append(result[len(result)-1].Categories, currentCategory)
	}

	return result
}

func MinePage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Columns []*BookmarkColumn
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	bookmarks, err := queries.GetBookmarksForUser(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("error getBookmarksForUser: %s", err)
			http.Error(w, "ERROR", 500)
			return
		}
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Columns:  preprocessBookmarks(bookmarks.List.String),
	}
	bookmarksCustomIndex(data.CoreData)

	if err := templates.RenderTemplate(w, "minePage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
