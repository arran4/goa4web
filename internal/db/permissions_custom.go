package db

import (
	"context"
	"database/sql"
	"strings"
)

// PermissionWithUser combines a permission with the associated user's info.
type PermissionWithUser struct {
	IduserRoles  int32
	UsersIdusers int32
	Role         sql.NullString
	Username     sql.NullString
	Email        sql.NullString
}

// GetPermissionsWithUsers returns all permissions joined with user details.
func (q *Queries) GetPermissionsWithUsers(ctx context.Context, user sql.NullString) ([]*PermissionWithUser, error) {
	query := `SELECT ur.iduser_roles, ur.users_idusers, r.name, u.username, u.email FROM user_roles ur JOIN users u ON u.idusers = ur.users_idusers JOIN roles r ON ur.role_id = r.id`
	var args []interface{}
	var cond []string
	if user.Valid {
		cond = append(cond, "u.username = ?")
		args = append(args, user.String)
	}
	if len(cond) > 0 {
		query += " WHERE " + strings.Join(cond, " AND ")
	}
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PermissionWithUser
	for rows.Next() {
		var i PermissionWithUser
		if err := rows.Scan(&i.IduserRoles, &i.UsersIdusers, &i.Role, &i.Username, &i.Email); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// UpdatePermission modifies an existing user's role.
type UpdatePermissionParams struct {
	ID   int32
	Role string
}

func (q *Queries) UpdatePermission(ctx context.Context, arg UpdatePermissionParams) error {
	const query = `UPDATE user_roles SET role_id = (SELECT id FROM roles WHERE name = ?) WHERE iduser_roles = ?`
	_, err := q.db.ExecContext(ctx, query, arg.Role, arg.ID)
	return err
}
