package role_templates

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// ApplyLogger is used to record apply operations.
type ApplyLogger interface {
	Printf(string, ...any)
}

// ApplyRoles inserts or updates roles and grants based on the provided template roles.
func ApplyRoles(ctx context.Context, q *db.Queries, tx *sql.Tx, roles []RoleDef, now time.Time, logger ApplyLogger) error {
	if logger == nil {
		logger = log.Default()
	}
	for _, rDef := range roles {
		role, err := q.GetRoleByName(ctx, rDef.Name)
		var roleID int32
		if err != nil {
			if err == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, "INSERT INTO roles (name, can_login, is_admin, private_labels, public_profile_allowed_at) VALUES (?, ?, ?, ?, NOW())", rDef.Name, rDef.CanLogin, rDef.IsAdmin, rDef.CanLogin)
				if err != nil {
					return fmt.Errorf("create role %s: %w", rDef.Name, err)
				}
				id, err := res.LastInsertId()
				if err != nil {
					return fmt.Errorf("get last insert id: %w", err)
				}
				roleID = int32(id)
				logger.Printf("Created role %s (ID: %d)", rDef.Name, roleID)
			} else {
				return fmt.Errorf("get role %s: %w", rDef.Name, err)
			}
		} else {
			roleID = role.ID
			if err := q.AdminUpdateRole(ctx, db.AdminUpdateRoleParams{
				Name:                   rDef.Name,
				CanLogin:               rDef.CanLogin,
				IsAdmin:                rDef.IsAdmin,
				PrivateLabels:          rDef.CanLogin,
				PublicProfileAllowedAt: sql.NullTime{Time: now, Valid: true},
				ID:                     role.ID,
			}); err != nil {
				return fmt.Errorf("update role %s: %w", rDef.Name, err)
			}
			logger.Printf("Updated role %s (ID: %d)", rDef.Name, roleID)
		}

		if err := q.DeleteGrantsByRoleID(ctx, sql.NullInt32{Int32: roleID, Valid: true}); err != nil {
			return fmt.Errorf("delete grants for role %s: %w", rDef.Name, err)
		}

		for _, g := range rDef.Grants {
			err := q.CreateGrant(ctx, db.CreateGrantParams{
				RoleID:   sql.NullInt32{Int32: roleID, Valid: true},
				Section:  g.Section,
				Item:     sql.NullString{String: g.Item, Valid: g.Item != ""},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: g.ItemID, Valid: g.ItemID != 0},
				Action:   g.Action,
				Active:   true,
			})
			if err != nil {
				return fmt.Errorf("create grant for %s (%s/%s/%s): %w", rDef.Name, g.Section, g.Item, g.Action, err)
			}
		}
	}
	return nil
}
