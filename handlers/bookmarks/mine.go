package bookmarks

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
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
		*common.CoreData
		Columns      []*BookmarkColumn
		HasBookmarks bool
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	_ = session
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	bookmarks, err := cd.Bookmarks()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("error getBookmarksForUser: %s", err)
		http.Error(w, "ERROR", 500)
		return
	}

	var list string
	if bookmarks != nil {
		list = bookmarks.List.String
	}
	list = strings.TrimSpace(list)

	data := Data{
		CoreData:     cd,
		HasBookmarks: list != "",
	}
	if list != "" {
		data.Columns = preprocessBookmarks(list)
	}

	handlers.TemplateHandler(w, r, "minePage.gohtml", data)
}
