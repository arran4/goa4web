package blogs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

// AdminPage shows the blog administration index with a list of blogs.
func AdminPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset := cd.Offset()
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/blogs?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/blogs?offset=%d", offset-ps)
		cd.StartLink = "/admin/blogs?offset=0"
	}
	cd.PageTitle = "Blog Admin"

	type Data struct {
		Rows []*db.ListUsersWithRolesRow
	}
	data := Data{}

	queries := cd.Queries()
	if queries != nil {
		rows, err := queries.ListUsersWithRoles(r.Context())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
			return
		}
		for _, row := range rows {
			if row.Roles.Valid && (strings.Contains(row.Roles.String, "content writer") || strings.Contains(row.Roles.String, "administrator")) {
				data.Rows = append(data.Rows, row)
			}
		}
	}

	handlers.TemplateHandler(w, r, "blogsAdminPage.gohtml", data)
}
