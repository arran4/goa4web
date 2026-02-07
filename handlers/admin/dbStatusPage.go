package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminDBStatusPage struct {
	DBPool *sql.DB
}

func (p *AdminDBStatusPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Status"
	if p.DBPool == nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	currentVersion, err := dbstart.SchemaVersion(r.Context(), p.DBPool)
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
		SeedAllowed:     cd.HasAdminRole() && p.DBPool != nil,
		SeedTask:        string(TaskDBSeed),
	}
	AdminDBStatusPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminDBStatusPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Database Status", "/admin/db/status", &AdminPage{}
}

func (p *AdminDBStatusPage) PageTitle() string {
	return "Database Status"
}

var _ common.Page = (*AdminDBStatusPage)(nil)
var _ http.Handler = (*AdminDBStatusPage)(nil)

// AdminDBStatusPageTmpl renders the database status page.
const AdminDBStatusPageTmpl tasks.Template = "admin/dbStatusPage.gohtml"
