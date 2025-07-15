package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	db "github.com/arran4/goa4web/internal/db"

	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

// BloggerListPage shows all bloggers with their post counts.
func BloggerListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows     []*db.BloggerCountRow
		Search   string
		NextLink string
		PrevLink string
		PageSize int
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Search:   r.URL.Query().Get("search"),
		PageSize: common.GetPageSize(r),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	pageSize := common.GetPageSize(r)
	var rows []*db.BloggerCountRow
	var err error
	if data.Search != "" {
		rows, err = queries.SearchBloggers(r.Context(), db.SearchBloggersParams{
			ViewerID: data.UserID,
			Query:    data.Search,
			Limit:    int32(pageSize + 1),
			Offset:   int32(offset),
		})
	} else {
		rows, err = queries.ListBloggers(r.Context(), db.ListBloggersParams{
			ViewerID: data.UserID,
			Limit:    int32(pageSize + 1),
			Offset:   int32(offset),
		})
	}
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	data.Rows = rows

	base := "/blogs/bloggers"
	if data.Search != "" {
		base += "?search=" + url.QueryEscape(data.Search)
	}
	if hasMore {
		if strings.Contains(base, "?") {
			data.NextLink = fmt.Sprintf("%s&offset=%d", base, offset+pageSize)
		} else {
			data.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+pageSize)
		}
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: fmt.Sprintf("Next %d", pageSize),
			Link: data.NextLink,
		})
	}
	if offset > 0 {
		if strings.Contains(base, "?") {
			data.PrevLink = fmt.Sprintf("%s&offset=%d", base, offset-pageSize)
		} else {
			data.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-pageSize)
		}
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: data.PrevLink,
		})
	}

	if err := templates.RenderTemplate(w, "bloggerListPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
