package common

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// CreatePrivateTopicWithAccess creates a private topic and all participant
// access grants in one transaction. Participant rows are locked in stable order
// so concurrent invitations for the same user cannot create duplicate global
// private-forum discovery grants.
func (cd *CoreData) CreatePrivateTopicWithAccess(p CreatePrivateTopicParams) (topicID int32, err error) {
	if cd == nil || cd.queries == nil {
		return 0, fmt.Errorf("no queries")
	}
	if !cd.HasGrant(consts.PermissionSectionPrivateForum.String(), consts.PermissionItemTopic.String(), consts.PermissionActionCreate.String(), 0) {
		return 0, fmt.Errorf("permission denied")
	}

	queries, ok := cd.queries.(*db.Queries)
	if !ok {
		return 0, fmt.Errorf("private topic creation requires sqlc queries")
	}

	usernames := make([]string, 0, len(p.Participants))
	participantIDs := make([]int32, 0, len(p.Participants))
	seenIDs := make(map[int32]struct{}, len(p.Participants))
	for _, participant := range p.Participants {
		if participant.ID == 0 {
			return 0, fmt.Errorf("invalid user id")
		}
		name := participant.Username
		if name == "" {
			if u := cd.UserByID(participant.ID); u != nil {
				name = u.Username.String
			} else {
				return 0, fmt.Errorf("unknown user %d", participant.ID)
			}
		}
		usernames = append(usernames, name)
		if _, exists := seenIDs[participant.ID]; !exists {
			seenIDs[participant.ID] = struct{}{}
			participantIDs = append(participantIDs, participant.ID)
		}
	}
	sort.Slice(participantIDs, func(i, j int) bool { return participantIDs[i] < participantIDs[j] })

	title := p.Title
	description := p.Description
	if title == "" {
		title = fmt.Sprintf("%s%s", PrivateTopicDefaultTitlePrefix, joinPrivateTopicUsernames(usernames))
		if description == "" {
			description = title
		}
	}

	tx, err := queries.BeginTx(cd.ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin private topic transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := queries.WithTx(tx)

	for _, userID := range participantIDs {
		var lockedUserID int32
		if err := tx.QueryRowContext(cd.ctx, "SELECT idusers FROM users WHERE idusers = ? FOR UPDATE", userID).Scan(&lockedUserID); err != nil {
			return 0, fmt.Errorf("lock private topic participant %d: %w", userID, err)
		}
		if err := ensurePrivateForumTopicSeeGrant(cd.ctx, qtx, userID); err != nil {
			return 0, err
		}
	}

	tid, err := qtx.CreateForumTopicForPoster(cd.ctx, db.CreateForumTopicForPosterParams{
		PosterID:        p.CreatorID,
		ForumcategoryID: PrivateForumCategoryID,
		ForumLang:       sql.NullInt32{},
		Title:           sql.NullString{String: title, Valid: true},
		Description:     sql.NullString{String: description, Valid: true},
		Handler:         "private",
		Section:         consts.PermissionSectionPrivateForum.String(),
		GrantCategoryID: sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
		GranteeID:       sql.NullInt32{Int32: p.CreatorID, Valid: p.CreatorID != 0},
	})
	if err != nil {
		return 0, fmt.Errorf("create topic: %w", err)
	}
	if tid == 0 {
		return 0, fmt.Errorf("create topic returned 0")
	}
	topicID = int32(tid)

	for _, participant := range p.Participants {
		uid := participant.ID
		for _, action := range []consts.PermissionAction{
			consts.PermissionActionSee,
			consts.PermissionActionView,
			consts.PermissionActionPost,
			consts.PermissionActionReply,
		} {
			if _, err := qtx.SystemCreateGrant(cd.ctx, db.SystemCreateGrantParams{
				UserID:   sql.NullInt32{Int32: uid, Valid: true},
				RoleID:   sql.NullInt32{},
				Section:  consts.PermissionSectionPrivateForum.String(),
				Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				ItemRule: sql.NullString{},
				Action:   action.String(),
				Extra:    sql.NullString{},
			}); err != nil {
				return 0, fmt.Errorf("create %s grant: %w", action, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit private topic transaction: %w", err)
	}
	return topicID, nil
}

func ensurePrivateForumTopicSeeGrant(ctx context.Context, queries *db.Queries, userID int32) error {
	_, err := queries.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
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
	_, err = queries.SystemCreateGrant(ctx, db.SystemCreateGrantParams{
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

func joinPrivateTopicUsernames(usernames []string) string {
	result := ""
	for i, username := range usernames {
		if i > 0 {
			result += ", "
		}
		result += username
	}
	return result
}
