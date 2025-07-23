package user

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"sort"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/internal/db"
)

func adminUsersPermissionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows  []*db.GetPermissionsWithUsersRow
		Roles []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	rows, err := queries.GetPermissionsWithUsers(r.Context(), db.GetPermissionsWithUsersParams{Username: sql.NullString{}})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Username.String < rows[j].Username.String
	})
	data.Rows = rows
}
