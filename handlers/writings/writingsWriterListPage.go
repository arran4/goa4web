package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// WriterListPage shows all writers with their article counts.
func WriterListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows                []*db.WriterCountRow
		Search              string
		NextLink            string
		PrevLink            string
		PageSize            int
		CategoryBreadcrumbs []*db.WritingCategory
		CategoryId          int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	handlers.SetPageTitle(r, "Writers")
	data := Data{
		Search:     r.URL.Query().Get("search"),
		PageSize:   cd.PageSize(),
		CategoryId: 0,
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pageSize := cd.PageSize()
	rows, err := cd.Writers(r)
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

	base := "/writings/writers"
	if data.Search != "" {
		base += "?search=" + url.QueryEscape(data.Search)
	}
	if hasMore {
		if strings.Contains(base, "?") {
			data.NextLink = fmt.Sprintf("%s&offset=%d", base, offset+pageSize)
		} else {
			data.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+pageSize)
		}
		cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
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
		cd.CustomIndexItems = append(cd.CustomIndexItems, common.IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: data.PrevLink,
		})
	}

	handlers.TemplateHandler(w, r, "writerListPage.gohtml", data)
}
