package admin

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminDBStatusPage shows the database schema version and maintenance actions.
func (h *Handlers) AdminDBStatusPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Status"
	if h.DBPool == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	currentVersion, err := dbstart.SchemaVersion(r.Context(), h.DBPool)
	if err != nil {
		log.Printf("db status schema version: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("internal server error"))
		return
	}
	expectedVersion := handlers.ExpectedSchemaVersion
	data := struct {
		CurrentVersion  int
		ExpectedVersion int
		VersionMatches  bool
		SeedAllowed     bool
		SeedTask        string
	}{
		CurrentVersion:  currentVersion,
		ExpectedVersion: expectedVersion,
		VersionMatches:  currentVersion == expectedVersion,
		SeedAllowed:     cd.HasAdminRole() && h.DBPool != nil,
		SeedTask:        string(TaskDBSeed),
	}
	AdminDBStatusPageTmpl.Handle(w, r, data)
}

// AdminDBStatusPageTmpl renders the database status page.
const AdminDBStatusPageTmpl tasks.Template = "admin/dbStatusPage.gohtml"
