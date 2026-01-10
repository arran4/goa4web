package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminQueuePage(w http.ResponseWriter, r *http.Request) {
	type QueueRow struct {
		*db.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetailsRow
		Preview string
	}
	type Data struct {
		Queue    []*QueueRow
		Search   string
		User     string
		Category string
		Offset   int
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Approval Queue"
	data := Data{
		Search:   r.URL.Query().Get("search"),
		User:     r.URL.Query().Get("user"),
		Category: r.URL.Query().Get("category"),
		Offset:   offset,
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	queue, err := queries.GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails Error: %s", err)
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
	}

	var filtered []*QueueRow
	for _, q := range queue {
		if data.User != "" && !strings.EqualFold(q.Username.String, data.User) {
			continue
		}
		if data.Category != "" {
			if !q.CategoryID.Valid || strconv.Itoa(int(q.CategoryID.Int32)) != data.Category {
				continue
			}
		}
		if data.Search != "" {
			s := strings.ToLower(data.Search)
			if !strings.Contains(strings.ToLower(q.Title.String), s) &&
				!strings.Contains(strings.ToLower(q.Description.String), s) &&
				!strings.Contains(strings.ToLower(q.Url.String), s) {
				continue
			}
		}
		filtered = append(filtered, &QueueRow{q, FetchPageTitle(r.Context(), q.Url.String)})
	}

	pageSize := cd.PageSize()
	if data.Offset < 0 {
		data.Offset = 0
	}
	if data.Offset > len(filtered) {
		data.Offset = len(filtered)
	}
	end := data.Offset + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	data.Queue = filtered[data.Offset:end]

	baseURL := "/admin/linker/queue"
	qv := make(url.Values)
	if data.Search != "" {
		qv.Set("search", data.Search)
	}
	if data.User != "" {
		qv.Set("user", data.User)
	}
	if data.Category != "" {
		qv.Set("category", data.Category)
	}
	qv.Set("offset", strconv.Itoa(data.Offset+pageSize))
	cd.NextLink = baseURL + "?" + qv.Encode()
	if data.Offset > 0 {
		qv.Set("offset", strconv.Itoa(data.Offset-pageSize))
		cd.PrevLink = baseURL + "?" + qv.Encode()
	}

	LinkerAdminQueuePageTmpl.Handle(w, r, data)
}

const LinkerAdminQueuePageTmpl handlers.Page = "linker/adminQueuePage.gohtml"

func AdminQueueUpdateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	title := r.URL.Query().Get("title")
	URL := r.URL.Query().Get("URL")
	desc := r.URL.Query().Get("desc")
	category, _ := strconv.Atoi(r.URL.Query().Get("category"))
	if err := queries.AdminUpdateLinkerQueuedItem(r.Context(), db.AdminUpdateLinkerQueuedItemParams{
		CategoryID:  sql.NullInt32{Int32: int32(category), Valid: category != 0},
		Title:       sql.NullString{Valid: true, String: title},
		Url:         sql.NullString{Valid: true, String: URL},
		Description: sql.NullString{Valid: true, String: desc},
		ID:          int32(qid),
	}); err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
