package db

import (
	"context"
	"strconv"
)

// ResolveRoleName converts an identifier into a role name. If ident is a
// numeric ID the role name is looked up from the database, otherwise ident is
// returned unchanged.
func (q *Queries) ResolveRoleName(ctx context.Context, ident string) (string, error) {
	if ident == "" {
		return "", nil
	}
	if id, err := strconv.Atoi(ident); err == nil {
		var name string
		err = q.db.QueryRowContext(ctx, "SELECT name FROM roles WHERE id = ?", id).Scan(&name)
		if err != nil {
			return "", err
		}
		return name, nil
	}
	return ident, nil
}
