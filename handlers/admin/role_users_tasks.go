package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

type RoleUsersAllowTask struct{ tasks.TaskString }

var roleUsersAllowTask = &RoleUsersAllowTask{TaskString: TaskUsersAllow}
var _ tasks.Task = (*RoleUsersAllowTask)(nil)

func (RoleUsersAllowTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	roleID := cd.SelectedRoleID()
	if roleID == 0 {
		return fmt.Errorf("role id parse fail %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	usernamesInput := r.PostFormValue("usernames")
	if usernamesInput == "" {
		return fmt.Errorf("usernames input is required %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}

	// Split by comma or newline
	var usernames []string
	for _, part := range strings.Split(strings.ReplaceAll(usernamesInput, "\n", ","), ",") {
		username := strings.TrimSpace(part)
		if username != "" {
			usernames = append(usernames, username)
		}
	}

	var errs []error
	for _, username := range usernames {
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("Failed to lookup user %s: %v", username, err)
			errs = append(errs, fmt.Errorf("failed to lookup %s", username))
			continue
		}

		if err := queries.SystemCreateUserRoleByID(r.Context(), db.SystemCreateUserRoleByIDParams{
			UsersIdusers: u.Idusers,
			RoleID:       roleID,
		}); err != nil {
			log.Printf("Failed to add user %s to role %d: %v", username, roleID, err)
			errs = append(errs, fmt.Errorf("failed to add %s", username))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred: %v %w", errs, handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/role/%d", roleID)}
}

type RoleUsersDisallowTask struct{ tasks.TaskString }

var roleUsersDisallowTask = &RoleUsersDisallowTask{TaskString: TaskUsersDisallow}
var _ tasks.Task = (*RoleUsersDisallowTask)(nil)

func (RoleUsersDisallowTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	roleID := cd.SelectedRoleID()
	if roleID == 0 {
		return fmt.Errorf("role id parse fail %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	permidStr := r.PostFormValue("permid")
	permid, err := strconv.Atoi(permidStr)
	if err != nil {
		return fmt.Errorf("permid parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if err := queries.AdminDeleteUserRole(r.Context(), int32(permid)); err != nil {
		log.Printf("Failed to delete user role mapping %d: %v", permid, err)
		return fmt.Errorf("failed to delete mapping %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/role/%d", roleID)}
}
