package role_templates

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// RoleDiff describes differences between a template role and the database role.
type RoleDiff struct {
	Name            string
	Status          string
	PropertyChanges []string
	GrantsAdded     []string
	GrantsRemoved   []string
}

// BuildTemplateDiff compares a template's roles and grants against the database.
func BuildTemplateDiff(ctx context.Context, q db.Querier, tmpl TemplateDef) ([]RoleDiff, error) {
	diffs := make([]RoleDiff, 0, len(tmpl.Roles))
	for _, rDef := range tmpl.Roles {
		diff := RoleDiff{Name: rDef.Name}
		role, err := q.GetRoleByName(ctx, rDef.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				diff.Status = "new"
				diffs = append(diffs, diff)
				continue
			}
			return nil, fmt.Errorf("get role %s: %w", rDef.Name, err)
		}

		diff.Status = "existing"
		if role.CanLogin != rDef.CanLogin {
			diff.PropertyChanges = append(diff.PropertyChanges, fmt.Sprintf("CanLogin: %v → %v", role.CanLogin, rDef.CanLogin))
		}
		if role.IsAdmin != rDef.IsAdmin {
			diff.PropertyChanges = append(diff.PropertyChanges, fmt.Sprintf("IsAdmin: %v → %v", role.IsAdmin, rDef.IsAdmin))
		}

		currentGrants, err := q.GetGrantsByRoleID(ctx, sql.NullInt32{Int32: role.ID, Valid: true})
		if err != nil {
			return nil, err
		}
		currSet := make(map[string]bool)
		for _, g := range currentGrants {
			key := GrantKey(g.Section, g.Item.String, g.Action, g.ItemID.Int32)
			currSet[key] = true
		}

		for _, g := range rDef.Grants {
			key := GrantKey(g.Section, g.Item, g.Action, g.ItemID)
			if !currSet[key] {
				diff.GrantsAdded = append(diff.GrantsAdded, key)
			} else {
				currSet[key] = false
			}
		}
		for key, exists := range currSet {
			if exists {
				diff.GrantsRemoved = append(diff.GrantsRemoved, key)
			}
		}
		diffs = append(diffs, diff)
	}
	return diffs, nil
}

// GrantKey formats a grant into a display-friendly key.
func GrantKey(section, item, action string, itemID int32) string {
	if item == "" {
		item = "*"
	}
	return fmt.Sprintf("%s / %s / %s [%d]", section, item, action, itemID)
}
