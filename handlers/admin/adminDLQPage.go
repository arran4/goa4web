package admin

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/handlers"

	"github.com/arran4/goa4web/internal/db"
	dirdlq "github.com/arran4/goa4web/internal/dlq/dir"
	filedlq "github.com/arran4/goa4web/internal/dlq/file"
)

// DeleteDLQTask deletes entries from the dead letter queue.
type DeleteDLQTask struct{ tasks.TaskString }

var deleteDLQTask = &DeleteDLQTask{TaskString: TaskDelete}
var updateDLQTask = &DeleteDLQTask{TaskString: TaskUpdate}
var purgeDLQTask = &DeleteDLQTask{TaskString: TaskPurge}

// compile-time interface check so DeleteDLQTask is usable as a generic task.
var _ tasks.Task = (*DeleteDLQTask)(nil)
var _ tasks.AuditableTask = (*DeleteDLQTask)(nil)

func AdminDLQPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Dead Letter Queue"
	data := struct {
		Errors     []*db.DeadLetter
		EditError  *db.DeadLetter
		FileErrors []filedlq.Record
		FileErr    string
		FilePath   string
		FileSize   int64
		FileMod    string
		FileTail   []string
		DirErrors  []dirdlq.Record
		DirErr     string
		DirPath    string
		DirCount   int
		DirMod     string
		DBCount    int64
		DBLatest   string
		Providers  string
	}{
		Providers: cd.Config.DLQProvider,
	}

	names := strings.Split(cd.Config.DLQProvider, ",")
	for i, n := range names {
		names[i] = strings.TrimSpace(strings.ToLower(n))
	}
	queries := cd.Queries()
	for _, n := range names {
		switch n {
		case "db":
			if rows, err := queries.SystemListDeadLetters(r.Context(), 100); err == nil {
				data.Errors = rows
			} else {
				log.Printf("list dead letters: %v", err)
			}
			if c, err := queries.SystemCountDeadLetters(r.Context()); err == nil {
				data.DBCount = c
			}
			if lt, err := queries.SystemLatestDeadLetter(r.Context()); err == nil {
				if t, ok := lt.(time.Time); ok {
					data.DBLatest = t.Format(time.RFC3339)
				}
			}
			if editIDStr := r.URL.Query().Get("edit"); editIDStr != "" {
				if id, err := strconv.Atoi(editIDStr); err == nil {
					if errStr, err := queries.SystemGetDeadLetter(r.Context(), int32(id)); err == nil {
						data.EditError = errStr
					} else {
						log.Printf("get dead letter: %v", err)
					}
				}
			}
		case "file":
			data.FilePath = cd.Config.DLQFile
			if st, err := os.Stat(cd.Config.DLQFile); err == nil {
				data.FileSize = st.Size()
				data.FileMod = st.ModTime().Format(time.RFC3339)
			}
			if lines, err := filedlq.Tail(cd.Config.DLQFile, 10); err == nil {
				data.FileTail = lines
			}
			if recs, err := filedlq.List(cd.Config.DLQFile, 100); err == nil {
				data.FileErrors = recs
			} else {
				log.Printf("read dlq file: %v", err)
				data.FileErr = err.Error()
			}
		case "dir":
			data.DirPath = cd.Config.DLQFile
			if entries, err := os.ReadDir(cd.Config.DLQFile); err == nil {
				data.DirCount = len(entries)
				if st, err2 := os.Stat(cd.Config.DLQFile); err2 == nil {
					data.DirMod = st.ModTime().Format(time.RFC3339)
				}
			}
			if recs, err := dirdlq.List(cd.Config.DLQFile, 100); err == nil {
				data.DirErrors = recs
			} else {
				log.Printf("read dlq dir: %v", err)
				data.DirErr = err.Error()
			}
		}
	}

	AdminDLQPageTmpl.Handle(w, r, data)
}

const AdminDLQPageTmpl tasks.Template = "admin/dlqPage.gohtml"

func (DeleteDLQTask) Action(w http.ResponseWriter, r *http.Request) any {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	switch r.PostFormValue("task") {
	case string(TaskDelete):
		for _, idStr := range r.Form["id"] {
			if idStr == "" {
				continue
			}
			id, _ := strconv.Atoi(idStr)
			if err := queries.SystemDeleteDeadLetter(r.Context(), int32(id)); err != nil {
				return fmt.Errorf("delete error %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["DeletedErrorID"] = appendID(evt.Data["DeletedErrorID"], id)
				}
			}
		}
	case string(TaskPurge):
		before := r.PostFormValue("before")
		t := time.Now()
		if before != "" {
			if tt, err := time.Parse("2006-01-02", before); err == nil {
				t = tt
			}
		}
		if err := queries.SystemPurgeDeadLettersBefore(r.Context(), t); err != nil {
			return fmt.Errorf("purge errors %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["PurgeBefore"] = t.Format(time.RFC3339)
			}
		}
	case string(TaskUpdate):
		idStr := r.PostFormValue("id")
		message := r.PostFormValue("message")
		if idStr != "" {
			id, _ := strconv.Atoi(idStr)
			if err := queries.SystemUpdateDeadLetter(r.Context(), db.SystemUpdateDeadLetterParams{
				ID:      int32(id),
				Message: message,
			}); err != nil {
				return fmt.Errorf("update error %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
				if evt := cd.Event(); evt != nil {
					if evt.Data == nil {
						evt.Data = map[string]any{}
					}
					evt.Data["UpdatedErrorID"] = id
				}
			}
		}
	}
	return nil
}

// AuditRecord summarises dead letters being removed or purged.
func (DeleteDLQTask) AuditRecord(data map[string]any) string {
	if ids, ok := data["DeletedErrorID"].(string); ok && ids != "" {
		return "deleted dead letters " + ids
	}
	if before, ok := data["PurgeBefore"].(string); ok && before != "" {
		return "purged dead letters before " + before
	}
	if id, ok := data["UpdatedErrorID"].(int); ok {
		return fmt.Sprintf("updated dead letter %d", id)
	}
	return "modified dead letter queue"
}
