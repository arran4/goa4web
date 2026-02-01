package admin

import (
	"encoding/json"
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
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	dirdlq "github.com/arran4/goa4web/internal/dlq/dir"
	filedlq "github.com/arran4/goa4web/internal/dlq/file"
)

// DeleteDLQTask deletes entries from the dead letter queue.
type DeleteDLQTask struct{ tasks.TaskString }

var deleteDLQTask = &DeleteDLQTask{TaskString: TaskDelete}
var updateDLQTask = &DeleteDLQTask{TaskString: TaskUpdate}
var purgeDLQTask = &DeleteDLQTask{TaskString: TaskPurge}

// ReEnlistDLQTask re-enlists a failed task from the DLQ.
type ReEnlistDLQTask struct{ tasks.TaskString }

var reEnlistDLQTask = &ReEnlistDLQTask{TaskString: "reenlist"}

// compile-time interface check so DeleteDLQTask is usable as a generic task.
var _ tasks.Task = (*DeleteDLQTask)(nil)
var _ tasks.AuditableTask = (*DeleteDLQTask)(nil)

var _ tasks.Task = (*ReEnlistDLQTask)(nil)
var _ tasks.AuditableTask = (*ReEnlistDLQTask)(nil)

type DisplayError struct {
	ID       any
	Message  string
	Time     time.Time
	Size     int64
	Parsed   *dlq.Message
	Raw      string
	Provider string
}

func AdminDLQPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Dead Letter Queue"

	data := struct {
		Errors     []*DisplayError
		EditError  *db.DeadLetter
		FileErrors []*DisplayError
		FileErr    string
		FilePath   string
		FileSize   int64
		FileMod    string
		FileTail   []string
		DirErrors  []*DisplayError
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

	parse := func(msg string) *dlq.Message {
		if !strings.HasPrefix(msg, "{") {
			return nil
		}
		var m dlq.Message
		if err := json.Unmarshal([]byte(msg), &m); err != nil {
			return nil
		}
		return &m
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
				for _, r := range rows {
					data.Errors = append(data.Errors, &DisplayError{
						ID:       r.ID,
						Message:  r.Message,
						Time:     r.CreatedAt,
						Raw:      r.Message,
						Parsed:   parse(r.Message),
						Size:     int64(len(r.Message)),
						Provider: "db",
					})
				}
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
				for _, r := range recs {
					data.FileErrors = append(data.FileErrors, &DisplayError{
						ID:       "",
						Time:     r.Time,
						Message:  r.Message,
						Raw:      r.Message,
						Parsed:   parse(r.Message),
						Provider: "file",
					})
				}
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
				for _, r := range recs {
					data.DirErrors = append(data.DirErrors, &DisplayError{
						ID:       r.Name,
						Message:  r.Message,
						Size:     r.Size,
						Raw:      r.Message,
						Parsed:   parse(r.Message),
						Provider: "dir",
					})
				}
			} else {
				log.Printf("read dlq dir: %v", err)
				data.DirErr = err.Error()
			}
		}
	}

	AdminDLQPageTmpl.Handle(w, r, data)
}

const AdminDLQPageTmpl tasks.Template = "admin/dlqPage.gohtml"
const AdminDLQDetailsPageTmpl tasks.Template = "admin/dlqDetailsPage.gohtml"

func AdminDLQDetailsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Dead Letter Queue Details"
	queries := cd.Queries()

	vars := mux.Vars(r)
	provider := vars["provider"]
	idStr := vars["id"]

	if provider == "" || idStr == "" {
		handlers.RenderErrorPage(w, r, fmt.Errorf("missing provider or id"))
		return
	}

	cfg := *cd.Config
	cfg.DLQProvider = provider
	inst := cd.DLQReg.ProviderFromConfig(&cfg, queries)

	m, ok := inst.(dlq.Manageable)
	if !ok {
		handlers.RenderErrorPage(w, r, fmt.Errorf("provider not manageable"))
		return
	}

	msgContent, err := m.Get(r.Context(), idStr)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("get message failed: %w", err))
		return
	}

	data := &DisplayError{
		ID:       idStr,
		Message:  msgContent,
		Raw:      msgContent,
		Provider: provider,
		Size:     int64(len(msgContent)),
	}

	// Try parse
	if strings.HasPrefix(msgContent, "{") {
		var dMsg dlq.Message
		if err := json.Unmarshal([]byte(msgContent), &dMsg); err == nil {
			data.Parsed = &dMsg
		}
	}

	// Try get more info if DB
	if provider == "db" {
		if id, err := strconv.Atoi(idStr); err == nil {
			if dl, err := queries.SystemGetDeadLetter(r.Context(), int32(id)); err == nil {
				data.Time = dl.CreatedAt
			}
		}
	}

	AdminDLQDetailsPageTmpl.Handle(w, r, data)
}

func (DeleteDLQTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	provider := r.PostFormValue("provider")

	switch r.PostFormValue("task") {
	case string(TaskDelete):
		for _, idStr := range r.Form["id"] {
			if idStr == "" {
				continue
			}
			if provider == "" && cd.Config.DLQProvider == "db" {
				provider = "db" // default fallback
			}
			// Use registry to create the provider instance
			cfg := *cd.Config
			cfg.DLQProvider = provider
			inst := cd.DLQReg.ProviderFromConfig(&cfg, queries)
			if m, ok := inst.(dlq.Manageable); ok {
				if err := m.Delete(r.Context(), idStr); err != nil {
					log.Printf("dlq delete failed: %v", err)
				} else {
					if evt := cd.Event(); evt != nil {
						if evt.Data == nil {
							evt.Data = map[string]any{}
						}
						evt.Data["DeletedErrorID"] = appendIDAny(evt.Data["DeletedErrorID"], idStr)
					}
				}
			} else {
				// Fallback for db delete if not manageable (should not happen with new db code)
				// or if provider logic fails.
				// Preserving old logic for backward compatibility/safety if provider is DB but registry failed?
				// Actually, if registry returns something not manageable, we can't delete easily unless we hardcode.
				// But we updated db provider to be manageable.
				if provider == "db" || provider == "" {
					id, _ := strconv.Atoi(idStr)
					if err := queries.SystemDeleteDeadLetter(r.Context(), int32(id)); err != nil {
						return fmt.Errorf("delete error %w", handlers.ErrRedirectOnSamePageHandler(err))
					}
					if evt := cd.Event(); evt != nil {
						if evt.Data == nil {
							evt.Data = map[string]any{}
						}
						evt.Data["DeletedErrorID"] = appendID(evt.Data["DeletedErrorID"], id)
					}
				} else {
					log.Printf("dlq provider %s is not manageable or unknown", provider)
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
		if provider == "dir" {
			// Not implemented safely yet
			return nil
		}
		if provider == "db" || provider == "" {
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

func (ReEnlistDLQTask) Action(w http.ResponseWriter, r *http.Request) any {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	provider := r.PostFormValue("provider")
	if provider == "" && cd.Config.DLQProvider == "db" {
		provider = "db"
	}

	for _, idStr := range r.Form["id"] {
		if idStr == "" {
			continue
		}
		var msgContent string

		cfg := *cd.Config
		cfg.DLQProvider = provider
		inst := cd.DLQReg.ProviderFromConfig(&cfg, cd.Queries())
		if m, ok := inst.(dlq.Manageable); ok {
			var err error
			msgContent, err = m.Get(r.Context(), idStr)
			if err != nil {
				return fmt.Errorf("get message %s error: %w", idStr, err)
			}
		} else {
			return fmt.Errorf("dlq provider %s is not manageable", provider)
		}

		var dlqMsg dlq.Message
		if err := json.Unmarshal([]byte(msgContent), &dlqMsg); err != nil {
			return fmt.Errorf("parse message json: %w", err)
		}

		if dlqMsg.Event == nil {
			return fmt.Errorf("no event in message")
		}
		evt := *dlqMsg.Event

		// Find Task
		if dlqMsg.TaskName != "" && cd.TasksReg != nil {
			found := false
			for _, t := range cd.TasksReg.Registered() {
				if t.Name() == dlqMsg.TaskName {
					if tt, ok := t.(tasks.Task); ok {
						evt.Task = tt
						found = true
					}
					break
				}
			}
			if !found {
				evt.Task = tasks.TaskString(dlqMsg.TaskName)
			}
		} else if dlqMsg.TaskName != "" {
			evt.Task = tasks.TaskString(dlqMsg.TaskName)
		}

		if err := cd.Publish(evt); err != nil {
			return fmt.Errorf("publish event: %w", err)
		}
	}

	return nil
}

// AuditRecord summarises dead letters being re-enlisted.
func (ReEnlistDLQTask) AuditRecord(data map[string]any) string {
	return "re-enlisted dead letter"
}

func appendIDAny(current any, newID string) string {
	var s string
	if current != nil {
		s = fmt.Sprint(current)
	}
	if s != "" {
		s += ", "
	}
	return s + newID
}
