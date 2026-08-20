package admin

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// TaskDBSeed applies the default seed data to the database.
const TaskDBSeed tasks.TaskString = "Seed database"

// DBSeedTask runs the database seed SQL.
type DBSeedTask struct {
	tasks.TaskString
	DBPool *sql.DB
}

// NewDBSeedTask creates a DB seed task bound to the admin DB pool.
func (h *Handlers) NewDBSeedTask() *DBSeedTask {
	return &DBSeedTask{TaskString: TaskDBSeed, DBPool: h.DBPool}
}

var _ tasks.Task = (*DBSeedTask)(nil)
var _ tasks.TaskMatcher = (*DBSeedTask)(nil)

// Matcher restricts seed tasks to administrators.
func (t *DBSeedTask) Matcher() mux.MatcherFunc {
	taskM := tasks.HasTask(string(TaskDBSeed))
	adminM := handlers.RequiredAccess("administrator")
	return func(r *http.Request, m *mux.RouteMatch) bool {
		return taskM(r, m) && adminM(r, m)
	}
}

// Action applies the embedded seed SQL to the database.
func (t *DBSeedTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd == nil || !cd.HasAdminRole() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: "/admin/db/status",
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", err)
	}
	if r.PostFormValue("confirm") != "yes" {
		data.Errors = []string{"Confirmation is required to apply seed data."}
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	if t.DBPool == nil {
		return fmt.Errorf("database not available")
	}
	if err := runSQLStatements(r.Context(), t.DBPool, strings.NewReader(string(database.SeedSQL))); err != nil {
		data.Errors = []string{fmt.Sprintf("Failed to apply seed data: %v", err)}
		return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
	}
	data.Messages = []string{"Seed data applied successfully."}
	if cd != nil {
		log.Printf("database seed applied by user %d", cd.UserID)
	}
	return handlers.TemplateWithDataHandler(handlers.TemplateRunTaskPage, data)
}

func runSQLStatements(ctx context.Context, db *sql.DB, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var stmt strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") || line == "" {
			continue
		}
		stmt.WriteString(line)
		if strings.HasSuffix(line, ";") {
			sqlStmt := strings.TrimSuffix(stmt.String(), ";")
			if _, err := db.ExecContext(ctx, sqlStmt); err != nil {
				return fmt.Errorf("executing statement %q: %w", sqlStmt, err)
			}
			stmt.Reset()
		} else {
			stmt.WriteString(" ")
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if s := strings.TrimSpace(stmt.String()); s != "" {
		if _, err := db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("executing statement %q: %w", s, err)
		}
	}
	return nil
}
