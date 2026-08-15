package common

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// EnsurePrivateForumTopicSeeGrant ensures a user can enter and list the private
// forum area. The private forum middleware checks topic/see without an item ID,
// so a topic-scoped see grant is not sufficient.
func (cd *CoreData) EnsurePrivateForumTopicSeeGrant(userID int32) error {
	if cd == nil || cd.queries == nil {
		return fmt.Errorf("no queries")
	}
	if userID == 0 {
		return fmt.Errorf("invalid user id")
	}

	_, err := cd.queries.SystemCheckGrant(cd.ctx, db.SystemCheckGrantParams{
		ViewerID: userID,
		Section:  consts.PermissionSectionPrivateForum.String(),
		Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		Action:   consts.PermissionActionSee.String(),
		ItemID:   sql.NullInt32{},
		UserID:   sql.NullInt32{Int32: userID, Valid: true},
	})
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check private forum see grant: %w", err)
	}

	_, err = cd.queries.SystemCreateGrant(cd.ctx, db.SystemCreateGrantParams{
		UserID:   sql.NullInt32{Int32: userID, Valid: true},
		RoleID:   sql.NullInt32{},
		Section:  consts.PermissionSectionPrivateForum.String(),
		Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{},
		ItemRule: sql.NullString{},
		Action:   consts.PermissionActionSee.String(),
		Extra:    sql.NullString{},
	})
	if err != nil {
		return fmt.Errorf("create private forum see grant: %w", err)
	}
	return nil
}
