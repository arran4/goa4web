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

	var usernames []string
	for _, part := range strings.Split(strings.ReplaceAll(usernamesInput, "\n", ","), ",") {
		username := strings.TrimSpace(part)
		if username != "" {
			usernames = append(usernames, username)
		}
	}

	var errs []string
	var failedUsernames []string
	for _, username := range usernames {
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{Valid: true, String: username})
		if err != nil {
			log.Printf("Failed to lookup user %s: %v", username, err)
			errs = append(errs, fmt.Sprintf("Failed to lookup user '%s'", username))
			failedUsernames = append(failedUsernames, username)
			continue
		}

		if err := queries.SystemCreateUserRoleByID(r.Context(), db.SystemCreateUserRoleByIDParams{
			UsersIdusers: u.Idusers,
			RoleID:       roleID,
		}); err != nil {
			log.Printf("Failed to add user %s to role %d: %v", username, roleID, err)
			errs = append(errs, fmt.Sprintf("Failed to add user '%s'", username))
			failedUsernames = append(failedUsernames, username)
		}
	}

	if len(errs) > 0 {
		role, err := queries.AdminGetRoleByID(r.Context(), roleID)
		if err != nil {
			return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/role/%d", roleID)}
		}

		id := roleID
		emailRows, _ := queries.GetVerifiedUserEmails(r.Context())
		emailsByUser := make(map[int32][]string)
		for _, row := range emailRows {
			emailsByUser[row.UserID] = append(emailsByUser[row.UserID], row.Email)
		}

		users, _ := queries.AdminListUsersByRoleID(r.Context(), id)
		roleUsers := make([]*roleUser, 0, len(users))
		for _, u := range users {
			ru := &roleUser{ID: u.Idusers, User: u.Username, UserID: u.Idusers, IduserRoles: u.IduserRoles}
			if emails, ok := emailsByUser[u.Idusers]; ok {
				ru.Email = emails
			}
			roleUsers = append(roleUsers, ru)
		}

		groups, _ := buildGrantGroups(r.Context(), cd, id)

		data := struct {
			Role        *db.Role
			Users       []*roleUser
			GrantGroups []GrantGroup
			Errors      []string
			Usernames   string
		}{
			Role:        role,
			Users:       roleUsers,
			GrantGroups: groups,
			Errors:      errs,
			Usernames:   strings.Join(failedUsernames, "\n"),
		}
		return handlers.TemplateWithDataHandler(AdminRolePageTmpl, data)
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
