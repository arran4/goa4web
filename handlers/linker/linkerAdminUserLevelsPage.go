package linker

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

func AdminUserLevelsPage(w http.ResponseWriter, r *http.Request) {
	type PermissionUser struct {
		*db.GetUserRolesRow
		Username sql.NullString
		Email    sql.NullString
	}

	type Data struct {
		*common.CoreData
		UserLevels []*PermissionUser
		Search     string
		Roles      []*db.Role
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*common.CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}
	rows, err := queries.GetUserRoles(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getUsersPermissions Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	var perms []*PermissionUser
	for _, p := range rows {
		row, err := queries.GetUserById(r.Context(), p.UsersIdusers)
		if err != nil {
			log.Printf("GetUserById Error: %s", err)
			continue
		}
		perms = append(perms, &PermissionUser{GetUserRolesRow: p, Username: row.Username, Email: row.Email})
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*PermissionUser
		for _, row := range perms {
			if strings.Contains(strings.ToLower(row.Username.String), q) {
				filtered = append(filtered, row)
			}
		}
		perms = filtered
	}
	data.UserLevels = perms

	handlers.TemplateHandler(w, r, "adminUserLevelsPage.gohtml", data)
}

type userAllowTask struct{ tasks.BasicTaskEvent }

var UserAllowTask = userAllowTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskUserAllow,
		Match:     tasks.HasTask(TaskUserAllow),
	},
}

func (userAllowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	usernames := r.PostFormValue("usernames")
	level := r.PostFormValue("role")
	fields := strings.FieldsFunc(usernames, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	for _, username := range fields {
		if username == "" {
			continue
		}
		u, err := queries.GetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("GetUserByUsername Error: %s", err)
			continue
		}
		if err := queries.CreateUserRole(r.Context(), db.CreateUserRoleParams{
			UsersIdusers: u.Idusers,
			Name:         level,
		}); err != nil {
			log.Printf("permissionUserAllow Error: %s", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

type userDisallowTask struct{ tasks.BasicTaskEvent }

var UserDisallowTask = userDisallowTask{
	BasicTaskEvent: tasks.BasicTaskEvent{
		EventName: TaskUserDisallow,
		Match:     tasks.HasTask(TaskUserDisallow),
	},
}

func (userDisallowTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	r.ParseForm()
	ids := r.Form["permids"]
	if len(ids) == 0 {
		if id := r.PostFormValue("permid"); id != "" {
			ids = append(ids, id)
		}
	}
	for _, idStr := range ids {
		permid, _ := strconv.Atoi(idStr)
		if err := queries.DeleteUserRole(r.Context(), int32(permid)); err != nil {
			log.Printf("permissionUserDisallow Error: %s", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}
