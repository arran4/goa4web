package common

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/db"
	"log"
	"slices"
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

	// Topic access tracking
	// UserID -> TopicID -> Action -> bool
	userTopicAccess := make(map[int32]map[int32]map[string]bool)

	// Thread access tracking
	// UserID -> ThreadID -> Action -> bool
	userThreadAccess := make(map[int32]map[int32]map[string]bool)

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

		// Rule 2: Must specify an item_id
		if !grant.ItemID.Valid {
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

		// Rule 2.5: Clean up legacy 'edit' topic grants
		if grant.Section == "privateforum" && grant.Item.String == "topic" && grant.Action == "edit" {
			inconsistencies = append(inconsistencies, PrivateForumInconsistency{
				ID:        fmt.Sprintf("delete-%d", grant.ID),
				GrantID:   grant.ID,
				Section:   grant.Section,
				Item:      grant.Item.String,
				Action:    grant.Action,
				ItemID:    grant.ItemID.Int32,
				RoleName:  roleName,
				UserID:    userID,
				Username:  username,
				Issue:     "Legacy 'edit' action on topic no longer supported",
				FixAction: "Delete grant",
			})
			continue
		}

		// Track user access
		if grant.UserID.Valid {
			if grant.Section == "privateforum" && grant.Item.String == "topic" {
				if userTopicAccess[userID] == nil {
					userTopicAccess[userID] = make(map[int32]map[string]bool)
				}
				if userTopicAccess[userID][grant.ItemID.Int32] == nil {
					userTopicAccess[userID][grant.ItemID.Int32] = make(map[string]bool)
				}
				userTopicAccess[userID][grant.ItemID.Int32][grant.Action] = true
			} else if grant.Section == "privateforum_thread" && grant.Item.String == "thread" {
				if userThreadAccess[userID] == nil {
					userThreadAccess[userID] = make(map[int32]map[string]bool)
				}
				if userThreadAccess[userID][grant.ItemID.Int32] == nil {
					userThreadAccess[userID][grant.ItemID.Int32] = make(map[string]bool)
				}
				userThreadAccess[userID][grant.ItemID.Int32][grant.Action] = true
			}
		}
	}

	// Fetch all threads
	threads, err := cd.queries.AdminListAllPrivateForumThreads(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all private forum threads: %w", err)
	}

	threadToTopic := make(map[int32]int32)
	topicToThreads := make(map[int32][]int32)
	for _, thread := range threads {
		threadToTopic[thread.Idforumthread] = thread.Idforumtopic
		topicToThreads[thread.Idforumtopic] = append(topicToThreads[thread.Idforumtopic], thread.Idforumthread)
	}

	// Rule 3: Check if user has thread access but NO topic access
	for _, grant := range grants {
		if !grant.UserID.Valid || grant.Section != "privateforum_thread" || grant.Item.String != "thread" || !grant.ItemID.Valid {
			continue
		}

		userID := grant.UserID.Int32
		threadID := grant.ItemID.Int32
		topicID, exists := threadToTopic[threadID]

		// Check if user has thread access but NO topic access for the same action
		// Also clean up any legacy 'edit' thread grants even if they have the topic grant
		if exists && (!userTopicAccess[userID][topicID][grant.Action] || grant.Action == "edit") {
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
				Issue:     fmt.Sprintf("User has access to thread %d but not its parent topic %d", threadID, topicID),
				FixAction: "Delete grant",
			})
		}
	}

	// Rule 4: User has topic access but is missing access to a thread in that topic
	// We need username mapping
	usernameMap := make(map[int32]string)
	for _, grant := range grants {
		if grant.UserID.Valid {
			usernameMap[grant.UserID.Int32] = grant.Username.String
		}
	}

	for userID, topics := range userTopicAccess {
		for topicID, actions := range topics {
			for _, threadID := range topicToThreads[topicID] {
				for action := range actions {
					if !userThreadAccess[userID][threadID][action] {
						inconsistencies = append(inconsistencies, PrivateForumInconsistency{
							ID:        fmt.Sprintf("create-%d-thread-%d-%s", userID, threadID, action),
							Section:   "privateforum_thread",
							Item:      "thread",
							Action:    action,
							ItemID:    threadID,
							UserID:    userID,
							Username:  usernameMap[userID],
							Issue:     fmt.Sprintf("User has %s access to topic %d but missing it for thread %d", action, topicID, threadID),
							FixAction: "Create missing thread grant",
						})
					}
				}
			}
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
				} else { // Create operation
					log.Printf("Fixing inconsistency: Creating grant for user %d thread %d (%s)", inconsistency.UserID, inconsistency.ItemID, inconsistency.Issue)

					_, err := cd.queries.SystemCreateGrant(ctx, db.SystemCreateGrantParams{
						Section:  inconsistency.Section,
						Item:     sql.NullString{String: inconsistency.Item, Valid: true},
						Action:   inconsistency.Action,
						UserID:   sql.NullInt32{Int32: inconsistency.UserID, Valid: true},
						RoleID:   sql.NullInt32{},
						ItemID:   sql.NullInt32{Int32: inconsistency.ItemID, Valid: true},
						ItemRule: sql.NullString{},
						RuleType: "allow",
					})

					if err != nil {
						log.Printf("Error creating missing grant: %v", err)
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
