package admincommon

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/arran4/goa4web/internal/db"
)

// UserRoleInfo aggregates a user's username, verified emails, and granted roles.
type UserRoleInfo struct {
	ID       int32
	Username sql.NullString
	Emails   []string
	Roles    []string
}

// LoadUserRoleInfo collects all users, their verified email addresses, and their roles.
// An optional roleFilter limits which roles are included for each user.
func LoadUserRoleInfo(ctx context.Context, queries db.Querier, roleFilter func(string) bool) ([]UserRoleInfo, error) {
	users, err := queries.AdminListAllUsers(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	userMap := make(map[int32]*UserRoleInfo, len(users))
	for _, u := range users {
		userMap[u.Idusers] = &UserRoleInfo{ID: u.Idusers, Username: u.Username}
	}

	emailRows, err := queries.GetVerifiedUserEmails(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	for _, row := range emailRows {
		u := userMap[row.UserID]
		if u == nil {
			u = &UserRoleInfo{ID: row.UserID}
			userMap[row.UserID] = u
		}
		duplicate := false
		for _, existing := range u.Emails {
			if existing == row.Email {
				duplicate = true
				break
			}
		}
		if !duplicate {
			u.Emails = append(u.Emails, row.Email)
		}
	}

	rows, err := queries.GetUserRoles(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	for _, row := range rows {
		if roleFilter != nil && !roleFilter(row.Role) {
			continue
		}
		u := userMap[row.UsersIdusers]
		if u == nil {
			u = &UserRoleInfo{ID: row.UsersIdusers, Username: row.Username}
			userMap[row.UsersIdusers] = u
		}
		roleDuplicate := false
		for _, role := range u.Roles {
			if role == row.Role {
				roleDuplicate = true
				break
			}
		}
		if !roleDuplicate {
			u.Roles = append(u.Roles, row.Role)
		}
	}

	result := make([]UserRoleInfo, 0, len(userMap))
	for _, u := range userMap {
		result = append(result, *u)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Username.String < result[j].Username.String
	})

	return result, nil
}
