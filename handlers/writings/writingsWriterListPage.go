package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// WriterListPage shows all writers with their article counts.
func WriterListPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
		Rows                []*db.WriterCountRow
		Search              string
		NextLink            string
		PrevLink            string
		PageSize            int
		IsAdmin             bool
		CategoryBreadcrumbs []*db.WritingCategory
		CategoryId          int32
	}

	data := Data{
		CoreData:   r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
		Search:     r.URL.Query().Get("search"),
		PageSize:   common.GetPageSize(r),
		IsAdmin:    false,
		CategoryId: 0,
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pageSize := common.GetPageSize(r)
	rows, err := data.CoreData.Writers(r)
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
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
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
		data.CustomIndexItems = append(data.CustomIndexItems, corecommon.IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: data.PrevLink,
		})
	}

	common.TemplateHandler(w, r, "writerListPage.gohtml", data)
}
