package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/tasks"
)

type AdminDBSchemaPage struct{}

func (p *AdminDBSchemaPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Schema"

	data := struct {
		Schema string
	}{
		Schema: string(database.SchemaMySQL),
	}
	AdminDBSchemaPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminDBSchemaPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Database Schema", "/admin/db/schema", &AdminDBStatusPage{}
}

func (p *AdminDBSchemaPage) PageTitle() string {
	return "Database Schema"
}

var _ common.Page = (*AdminDBSchemaPage)(nil)
var _ http.Handler = (*AdminDBSchemaPage)(nil)

// AdminDBSchemaPageTmpl renders the database schema page.
const AdminDBSchemaPageTmpl tasks.Template = "admin/dbSchemaPage.gohtml"
