package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
		*common.CoreData
		Queue    []*QueueRow
		Search   string
		User     string
		Category string
		Offset   int
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData: cd,
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	var filtered []*QueueRow
	for _, q := range queue {
		if data.User != "" && !strings.EqualFold(q.Username.String, data.User) {
			continue
		}
		if data.Category != "" && strconv.Itoa(int(q.LinkerCategoryID)) != data.Category {
			continue
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
	next := qv.Encode()
	if next != "" {
		next = "?" + next + "&offset=%d"
	} else {
		next = "?offset=%d"
	}
	data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
		Name: fmt.Sprintf("Next %d", pageSize),
		Link: baseURL + fmt.Sprintf(next, data.Offset+pageSize),
	})
	if data.Offset > 0 {
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: baseURL + fmt.Sprintf(next, data.Offset-pageSize),
		})
	}

	handlers.TemplateHandler(w, r, "adminQueuePage.gohtml", data)
}

func AdminQueueUpdateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	title := r.URL.Query().Get("title")
	URL := r.URL.Query().Get("URL")
	desc := r.URL.Query().Get("desc")
	category, _ := strconv.Atoi(r.URL.Query().Get("category"))
	if err := queries.UpdateLinkerQueuedItem(r.Context(), db.UpdateLinkerQueuedItemParams{
		LinkerCategoryID: int32(category),
		Title:            sql.NullString{Valid: true, String: title},
		Url:              sql.NullString{Valid: true, String: URL},
		Description:      sql.NullString{Valid: true, String: desc},
		Idlinkerqueue:    int32(qid),
	}); err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
