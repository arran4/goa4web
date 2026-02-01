package admin

import (
	"io/fs"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/migrations"
)

// AdminDBMigrationsPage shows the database migrations.
func (h *Handlers) AdminDBMigrationsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Database Migrations"

	files, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	var filenames []string
	for _, f := range files {
		if !f.IsDir() {
			filenames = append(filenames, f.Name())
		}
	}
	sort.Strings(filenames)

	selectedFile := r.FormValue("name")
	var content string

	if selectedFile != "" {
		valid := false
		for _, f := range filenames {
			if f == selectedFile {
				valid = true
				break
			}
		}
		if valid {
			b, err := fs.ReadFile(migrations.FS, selectedFile)
			if err == nil {
				content = string(b)
			}
		}
	}

	data := struct {
		Files        []string
		SelectedFile string
		Content      string
	}{
		Files:        filenames,
		SelectedFile: selectedFile,
		Content:      content,
	}
	AdminDBMigrationsPageTmpl.Handle(w, r, data)
}

// AdminDBMigrationsPageTmpl renders the database migrations page.
const AdminDBMigrationsPageTmpl tasks.Template = "admin/dbMigrationsPage.gohtml"
