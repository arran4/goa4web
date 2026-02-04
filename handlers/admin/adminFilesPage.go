package admin

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

func AdminFilesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Managed Files"
	queries := cd.Queries()

	pageSize := cd.PageSize()
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	type Data struct {
		Entries  []*db.AdminListAllImagePostsRow
		Total    int64
		PageSize int
	}

	total, err := queries.AdminCountAllImagePosts(r.Context())
	if err != nil {
		log.Printf("count images: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	rows, err := queries.AdminListAllImagePosts(r.Context(), db.AdminListAllImagePostsParams{
		Limit:  int32(pageSize + 1),
		Offset: int32(offset),
	})
	if err != nil {
		log.Printf("list images: %v", err)
		handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
		return
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	data := Data{
		Entries:  rows,
		Total:    total,
		PageSize: pageSize,
	}

	// Pagination Links
	params := url.Values{}
	for k, v := range r.URL.Query() {
		if k != "offset" {
			params[k] = v
		}
	}

	if hasMore {
		nextVals := url.Values{}
		for k, v := range params {
			nextVals[k] = v
		}
		nextVals.Set("offset", strconv.Itoa(offset+pageSize))
		cd.NextLink = r.URL.Path + "?" + nextVals.Encode()
	}
	if offset > 0 {
		prev := offset - pageSize
		if prev < 0 {
			prev = 0
		}
		prevVals := url.Values{}
		for k, v := range params {
			prevVals[k] = v
		}
		prevVals.Set("offset", strconv.Itoa(prev))
		cd.PrevLink = r.URL.Path + "?" + prevVals.Encode()
	}

	AdminFilesPageTmpl.Handle(w, r, data)
}

const AdminFilesPageTmpl tasks.Template = "admin/adminFilesPage.gohtml"
