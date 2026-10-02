package common

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/arran4/goa4web/core/consts"
)

// PrivateForumInconsistency represents a found inconsistency in private forum grants
type PrivateForumInconsistency struct {
	ID        string // A unique ID to identify the fix (since not all fixes are deleting a grant)
	GrantID   int32
	Section   string
	Item      string
	Action    string
	ItemID    int32
	RoleName  string
	UserID    int32
	Username  string
	Issue     string
	FixAction string
}

// CheckAndFixPrivateForumInconsistencies finds and optionally fixes private forum permission inconsistencies.
func (cd *CoreData) CheckAndFixPrivateForumInconsistencies(ctx context.Context, fixIDs []string, dryRun bool) ([]PrivateForumInconsistency, error) {
	var inconsistencies []PrivateForumInconsistency

	// Get all private forum grants
	grants, err := cd.queries.AdminListAllPrivateForumGrants(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting private forum grants: %w", err)
	}

	// Topic view is the parent eligibility boundary for fine-grained thread grants.
	// UserID -> TopicID -> bool
	userTopicViewAccess := make(map[int32]map[int32]bool)

	for _, grant := range grants {
		roleName := ""
		if grant.RoleName.Valid {
			roleName = grant.RoleName.String
		}
		username := ""
		if grant.Username.Valid {
			username = grant.Username.String
		}
		userID := int32(0)
		if grant.UserID.Valid {
			userID = grant.UserID.Int32
		}

		// Rule 1: No "anyone" access
		if !grant.RoleName.Valid && !grant.UserID.Valid {
			inconsistencies = append(inconsistencies, PrivateForumInconsistency{
				ID:        fmt.Sprintf("delete-%d", grant.ID),
				GrantID:   grant.ID,
				Section:   grant.Section,
				Item:      grant.Item.String,
				Action:    grant.Action,
				ItemID:    grant.ItemID.Int32,
				Issue:     "Grant allows 'anyone' access (no role, no user)",
				FixAction: "Delete grant",
			})
			continue
		}

		// Rule 2: Resource grants must specify an item_id.
		// Unscoped capability grants (e.g. edit, label on privateforum topic) are permitted without an item_id.
		if !grant.ItemID.Valid {
			if isPrivateForumCapabilityGrant(grant.Section, grant.Item.String, grant.Action) {
				continue
			}
			inconsistencies = append(inconsistencies, PrivateForumInconsistency{
				ID:        fmt.Sprintf("delete-%d", grant.ID),
				GrantID:   grant.ID,
				Section:   grant.Section,
				Item:      grant.Item.String,
				Action:    grant.Action,
				RoleName:  roleName,
				UserID:    userID,
				Username:  username,
				Issue:     "Grant has no item_id (access to all topics/threads)",
				FixAction: "Delete grant",
			})
			continue
		}

		// Track user access
		if grant.UserID.Valid {
			if grant.Section == consts.PermissionSectionPrivateForum.String() &&
				grant.Item.String == consts.PermissionItemTopic.String() &&
				grant.Action == consts.PermissionActionView.String() &&
				grant.ItemID.Valid {
				if userTopicViewAccess[userID] == nil {
					userTopicViewAccess[userID] = make(map[int32]bool)
				}
				userTopicViewAccess[userID][grant.ItemID.Int32] = true
			}
		}
	}

	// Fetch all threads
	threads, err := cd.queries.AdminListAllPrivateForumThreads(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all private forum threads: %w", err)
	}

	threadToTopic := make(map[int32]int32)
	for _, thread := range threads {
		threadToTopic[thread.Idforumthread] = thread.Idforumtopic
	}

	// Rule 3: Item-specific private-thread permissions require parent topic view access.
	for _, grant := range grants {
		if !grant.UserID.Valid || grant.Section != consts.PermissionSectionPrivateForumThread.String() || grant.Item.String != consts.PermissionItemThread.String() || !grant.ItemID.Valid {
			continue
		}

		userID := grant.UserID.Int32
		threadID := grant.ItemID.Int32
		topicID, exists := threadToTopic[threadID]

		if !exists {
			inconsistencies = append(inconsistencies, PrivateForumInconsistency{
				ID:        fmt.Sprintf("delete-%d", grant.ID),
				GrantID:   grant.ID,
				Section:   grant.Section,
				Item:      grant.Item.String,
				Action:    grant.Action,
				ItemID:    threadID,
				RoleName:  "",
				UserID:    userID,
				Username:  grant.Username.String,
				Issue:     fmt.Sprintf("Thread %d does not exist (orphaned grant)", threadID),
				FixAction: "Delete grant",
			})
			continue
		}

		if isPrivateForumThreadAction(grant.Action) {
			if !userTopicViewAccess[userID][topicID] {
				inconsistencies = append(inconsistencies, PrivateForumInconsistency{
					ID:        fmt.Sprintf("delete-%d", grant.ID),
					GrantID:   grant.ID,
					Section:   grant.Section,
					Item:      grant.Item.String,
					Action:    grant.Action,
					ItemID:    threadID,
					RoleName:  "",
					UserID:    userID,
					Username:  grant.Username.String,
					Issue:     fmt.Sprintf("User has %s access to thread %d without view access to parent topic %d", grant.Action, threadID, topicID),
					FixAction: "Delete grant",
				})
			}
		} else {
			inconsistencies = append(inconsistencies, PrivateForumInconsistency{
				ID:        fmt.Sprintf("delete-%d", grant.ID),
				GrantID:   grant.ID,
				Section:   grant.Section,
				Item:      grant.Item.String,
				Action:    grant.Action,
				ItemID:    threadID,
				RoleName:  "",
				UserID:    userID,
				Username:  grant.Username.String,
				Issue:     fmt.Sprintf("Unsupported action %s on private forum thread %d", grant.Action, threadID),
				FixAction: "Delete grant",
			})
		}
	}

	if !dryRun {
		for _, inconsistency := range inconsistencies {
			if slices.Contains(fixIDs, inconsistency.ID) {
				if inconsistency.GrantID > 0 { // Delete operation
					log.Printf("Fixing inconsistency: Deleting grant ID %d (%s)", inconsistency.GrantID, inconsistency.Issue)
					err := cd.queries.AdminDeleteGrant(ctx, inconsistency.GrantID)
					if err != nil {
						log.Printf("Error deleting grant %d: %v", inconsistency.GrantID, err)
					}
				}
			}
		}

		var processed []PrivateForumInconsistency
		for _, inconsistency := range inconsistencies {
			if slices.Contains(fixIDs, inconsistency.ID) {
				processed = append(processed, inconsistency)
			}
		}
		inconsistencies = processed
	}

	return inconsistencies, nil
}

func isPrivateForumCapabilityGrant(section, item, action string) bool {
	if section == consts.PermissionSectionPrivateForum.String() && item == consts.PermissionItemTopic.String() {
		switch action {
		case consts.PermissionActionEdit.String(),
			consts.PermissionActionLabel.String(),
			consts.PermissionActionCreate.String(),
			consts.PermissionActionSee.String(),
			consts.PermissionActionView.String(),
			consts.PermissionActionPost.String(),
			consts.PermissionActionReply.String():
			return true
		}
	}
	return false
}

func isPrivateForumThreadAction(action string) bool {
	switch action {
	case consts.PermissionActionView.String(),
		consts.PermissionActionReply.String(),
		consts.PermissionActionEdit.String(),
		consts.PermissionActionEditAny.String(),
		"append":
		return true
	default:
		return false
	}
}
