package goa4web

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/templates"
)

func blogsBloggersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows     []*BloggerCountRow
		Search   string
		NextLink string
		PrevLink string
		PageSize int
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Search:   r.URL.Query().Get("search"),
		PageSize: getPageSize(r),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	pageSize := getPageSize(r)
	var rows []*BloggerCountRow
	var err error
	if data.Search != "" {
		rows, err = queries.SearchBloggers(r.Context(), SearchBloggersParams{
			Query:  data.Search,
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
		})
	} else {
		rows, err = queries.ListBloggers(r.Context(), ListBloggersParams{
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
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

	CustomBlogIndex(data.CoreData, r)

	if err := templates.RenderTemplate(w, "bloggersPage.gohtml", data, NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
