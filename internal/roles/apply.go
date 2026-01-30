package roles

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/arran4/goa4web/internal/db"
)

// ApplyRoleGrants copies grants from a source role to a destination role.
func ApplyRoleGrants(ctx context.Context, sdb *sql.DB, queries db.Querier, srcRoleName string, destRoleName string) error {
	srcRole, err := queries.GetRoleByName(ctx, srcRoleName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Role %q not found in database. Searching embedded roles...", srcRoleName)
			fileName, findErr := FindEmbeddedRoleByName(srcRoleName)
			if findErr != nil {
				return fmt.Errorf("failed to get source role by name: %w (and failed to find embedded role: %v)", err, findErr)
			}
			log.Printf("Found embedded role file %q for role %q. Attempting to load...", fileName, srcRoleName)
			if loadErr := LoadRole(ctx, fileName, "", sdb); loadErr != nil {
				return fmt.Errorf("failed to load embedded role %q: %w", fileName, loadErr)
			}
			srcRole, err = queries.GetRoleByName(ctx, srcRoleName)
			if err != nil {
				return fmt.Errorf("failed to get source role by name after loading: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get source role by name: %w", err)
		}
	}

	destRole, err := queries.GetRoleByName(ctx, destRoleName)
	if err != nil {
		return fmt.Errorf("failed to get destination role by name: %w", err)
	}

	grants, err := queries.GetGrantsByRoleID(ctx, sql.NullInt32{Int32: srcRole.ID, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get grants for source role: %w", err)
	}

	log.Printf("Applying %d grants from %q to %q", len(grants), srcRoleName, destRoleName)
	for _, grant := range grants {
		params := db.CreateGrantParams{
			RoleID:   sql.NullInt32{Int32: destRole.ID, Valid: true},
			Section:  grant.Section,
			Item:     grant.Item,
			RuleType: grant.RuleType,
			ItemID:   grant.ItemID,
			ItemRule: grant.ItemRule,
			Action:   grant.Action,
			Extra:    grant.Extra,
			Active:   grant.Active,
		}
		if err := queries.CreateGrant(ctx, params); err != nil {
			return fmt.Errorf("failed to create grant: %w", err)
		}
	}

	log.Printf("Successfully applied grants from %q to %q.", srcRoleName, destRoleName)
	return nil
}
