package admin

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/tasks"
)

// AdminDBSchemaPage shows the database schema.
func (h *Handlers) AdminDBSchemaPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Schema"

	data := struct {
		Schema string
	}{
		Schema: string(database.SchemaMySQL),
	}
	AdminDBSchemaPageTmpl.Handle(w, r, data)
}

// AdminDBSchemaPageTmpl renders the database schema page.
const AdminDBSchemaPageTmpl tasks.Template = "admin/dbSchemaPage.gohtml"
