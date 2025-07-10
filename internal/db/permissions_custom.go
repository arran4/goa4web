package db

import (
	"context"
	"database/sql"
	"strings"
)

// PermissionWithUser combines a permission with the associated user's info.
type PermissionWithUser struct {
	Idpermissions int32
	UsersIdusers  int32
	Section       sql.NullString
	Role          sql.NullString
	Username      sql.NullString
	Email         sql.NullString
}

// GetPermissionsWithUsers returns all permissions joined with user details.
func (q *Queries) GetPermissionsWithUsers(ctx context.Context, user, section sql.NullString) ([]*PermissionWithUser, error) {
	query := `SELECT p.idpermissions, p.users_idusers, p.section, p.level, u.username, u.email FROM permissions p JOIN users u ON u.idusers = p.users_idusers`
	var args []interface{}
	var cond []string
	if user.Valid {
		cond = append(cond, "u.username = ?")
		args = append(args, user.String)
	}
	if section.Valid {
		cond = append(cond, "p.section = ?")
		args = append(args, section.String)
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
		if err := rows.Scan(&i.Idpermissions, &i.UsersIdusers, &i.Section, &i.Role, &i.Username, &i.Email); err != nil {
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

// UpdatePermission modifies an existing permission's section and role.
type UpdatePermissionParams struct {
	ID      int32
	Section sql.NullString
	Role    sql.NullString
}

func (q *Queries) UpdatePermission(ctx context.Context, arg UpdatePermissionParams) error {
	const query = `UPDATE permissions SET section = ?, role = ? WHERE idpermissions = ?`
	_, err := q.db.ExecContext(ctx, query, arg.Section, arg.Role, arg.ID)
	return err
}
