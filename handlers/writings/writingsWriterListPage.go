package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// WriterListPage shows all writers with their article counts.
func WriterListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Rows                []*db.ListWritersForListerRow
		Search              string
		PageSize            int
		CategoryBreadcrumbs []*db.WritingCategory
		CategoryId          int32
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Writers"
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
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
			cd.NextLink = fmt.Sprintf("%s&offset=%d", base, offset+pageSize)
		} else {
			cd.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+pageSize)
		}
	}
	if offset > 0 {
		if strings.Contains(base, "?") {
			cd.PrevLink = fmt.Sprintf("%s&offset=%d", base, offset-pageSize)
		} else {
			cd.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-pageSize)
		}
	}

	WritingsWriterListPageTmpl.Handle(w, r, data)
}

const WritingsWriterListPageTmpl handlers.Page = "writings/writerListPage.gohtml"
