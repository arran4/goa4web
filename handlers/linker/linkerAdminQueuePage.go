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

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	searchworker "github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/internal/tasks"
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
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Search:   r.URL.Query().Get("search"),
		User:     r.URL.Query().Get("user"),
		Category: r.URL.Query().Get("category"),
		Offset:   offset,
	}

	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)

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

	pageSize := handlers.GetPageSize(r)
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

type deleteTask struct{ tasks.TaskString }

var DeleteTask = &deleteTask{TaskString: TaskDelete}

func (deleteTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	if err := queries.DeleteLinkerQueuedItem(r.Context(), int32(qid)); err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func AdminQueueUpdateActionPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
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

type approveTask struct{ tasks.TaskString }

var ApproveTask = &approveTask{TaskString: TaskApprove}

func (approveTask) IndexType() string { return searchworker.TypeLinker }

func (approveTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = approveTask{}

func (approveTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	qid, _ := strconv.Atoi(r.URL.Query().Get("qid"))
	lid, err := queries.SelectInsertLInkerQueuedItemIntoLinkerByLinkerQueueId(r.Context(), int32(qid))
	if err != nil {
		log.Printf("updateLinkerQueuedItem Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(lid))
	if err != nil {
		log.Printf("getLinkerItemById Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	text := strings.Join([]string{link.Title.String, link.Description.String}, " ")
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeLinker, ID: int32(lid), Text: text}
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

type bulkDeleteTask struct{ tasks.TaskString }

var BulkDeleteTask = &bulkDeleteTask{TaskString: TaskBulkDelete}

func (bulkDeleteTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
	}
	for _, q := range r.Form["qid"] {
		id, _ := strconv.Atoi(q)
		if err := queries.DeleteLinkerQueuedItem(r.Context(), int32(id)); err != nil {
			log.Printf("deleteLinkerQueuedItem Error: %s", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

type bulkApproveTask struct{ tasks.TaskString }

var BulkApproveTask = &bulkApproveTask{TaskString: TaskBulkApprove}

func (bulkApproveTask) IndexType() string { return searchworker.TypeLinker }

func (bulkApproveTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = bulkApproveTask{}

func (bulkApproveTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
	}
	for _, q := range r.Form["qid"] {
		id, _ := strconv.Atoi(q)
		lid, err := queries.SelectInsertLInkerQueuedItemIntoLinkerByLinkerQueueId(r.Context(), int32(id))
		if err != nil {
			log.Printf("selectInsert Error: %s", err)
			continue
		}
		link, err := queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(r.Context(), int32(lid))
		if err != nil {
			log.Printf("getLinkerItemById Error: %s", err)
			continue
		}
		text := strings.Join([]string{link.Title.String, link.Description.String}, " ")
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeLinker, ID: int32(lid), Text: text}
			}
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
