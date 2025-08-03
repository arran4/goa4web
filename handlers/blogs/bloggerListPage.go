package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"

	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/handlers"
)

// BloggerListPage shows all bloggers with their post counts.
func BloggerListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows     []*db.ListBloggersForListerRow
		Search   string
		NextLink string
		PrevLink string
		PageSize int
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Bloggers"
	data := Data{
		CoreData: cd,
		Search:   r.URL.Query().Get("search"),
		PageSize: cd.PageSize(),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pageSize := cd.PageSize()
	rows, err := data.CoreData.Bloggers(r)
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
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
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
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: data.PrevLink,
		})
	}

	handlers.TemplateHandler(w, r, "bloggerListPage.gohtml", data)
}
