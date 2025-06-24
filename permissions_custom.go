package goa4web

import (
	"context"
	"database/sql"
)

// PermissionWithUser combines a permission with the associated user's info.
type PermissionWithUser struct {
	Idpermissions int32
	UsersIdusers  int32
	Section       sql.NullString
	Level         sql.NullString
	Username      sql.NullString
	Email         sql.NullString
}

// GetPermissionsWithUsers returns all permissions joined with user details.
func (q *Queries) GetPermissionsWithUsers(ctx context.Context) ([]*PermissionWithUser, error) {
	const query = `SELECT p.idpermissions, p.users_idusers, p.section, p.level, u.username, u.email
                    FROM permissions p JOIN users u ON u.idusers = p.users_idusers`
	rows, err := q.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PermissionWithUser
	for rows.Next() {
		var i PermissionWithUser
		if err := rows.Scan(&i.Idpermissions, &i.UsersIdusers, &i.Section, &i.Level, &i.Username, &i.Email); err != nil {
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

// UpdatePermission modifies an existing permission's section and level.
type UpdatePermissionParams struct {
	ID      int32
	Section sql.NullString
	Level   sql.NullString
}

func (q *Queries) UpdatePermission(ctx context.Context, arg UpdatePermissionParams) error {
	const query = `UPDATE permissions SET section = ?, level = ? WHERE idpermissions = ?`
	_, err := q.db.ExecContext(ctx, query, arg.Section, arg.Level, arg.ID)
	return err
}
