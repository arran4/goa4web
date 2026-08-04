package common

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// MergeGroup represents a group of topics being merged
type MergeGroup struct {
	PrimaryTopicID int32
	MergedTopicIDs []int32
	Participants   []string
}

// MergePrivateTopicsWithSameParticipants finds private forum topics with identical participants
// and merges their threads into a single primary topic. The remaining topics are deleted.
// It returns a list of MergeGroup detailing the operations, and an error if one occurred.
func (cd *CoreData) MergePrivateTopicsWithSameParticipants(ctx context.Context, dryRun bool) ([]MergeGroup, error) {
	topics, err := cd.queries.AdminListAllPrivateTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting private forum topics: %w", err)
	}

	topicParticipants := make(map[int32][]string)
	for _, topic := range topics {
		participants, err := cd.queries.AdminListPrivateTopicParticipantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
		if err != nil {
			log.Printf("error getting participants for topic %d: %v", topic.Idforumtopic, err)
			continue
		}
		var usernames []string
		for _, p := range participants {
			if p.Username.Valid {
				usernames = append(usernames, p.Username.String)
			}
		}
		sort.Strings(usernames)
		topicParticipants[topic.Idforumtopic] = usernames
	}

	// Find duplicate topics
	participantsToTopics := make(map[string][]int32)
	for topicID, participants := range topicParticipants {
		key := strings.Join(participants, ",")
		participantsToTopics[key] = append(participantsToTopics[key], topicID)
	}

	var mergeGroups []MergeGroup

	for key, topicIDs := range participantsToTopics {
		if len(topicIDs) <= 1 {
			continue
		}
		if key == "" {
			continue // Skip topics with no participants (should be handled by clean-empty)
		}

		slices.Sort(topicIDs)

		primaryTopicID := topicIDs[0]
		mergedIDs := topicIDs[1:]

		log.Printf("Merging topics into primary topic ID %d (Participants: %s)", primaryTopicID, key)

		for _, topicID := range mergedIDs {
			log.Printf("  - Merging from topic ID: %d", topicID)
			if !dryRun {
				// Move threads
				err := cd.queries.AdminMoveThreadsToTopic(ctx, db.AdminMoveThreadsToTopicParams{
					ForumtopicIdforumtopic:   primaryTopicID,
					ForumtopicIdforumtopic_2: topicID,
				})
				if err != nil {
					log.Printf("  - error moving threads from %d to %d: %v", topicID, primaryTopicID, err)
					continue
				}

				// The primary topic already has exact copies of grants for these participants.
				// We don't need to move them. We can safely delete the topic and let cascade delete or explicit delete remove the grants.

				// Delete old grants on the duplicated topic to avoid clutter.
				grants, err := cd.queries.AdminListGrantsByTopicID(ctx, sql.NullInt32{Int32: topicID, Valid: true})
				if err == nil {
					for _, grant := range grants {
						err = cd.queries.AdminDeleteGrant(ctx, grant.ID)
						if err != nil {
							log.Printf("  - error deleting grant %d on topic %d: %v", grant.ID, topicID, err)
						}
					}
				}

				// Delete old topic
				err = cd.queries.AdminDeleteForumTopic(ctx, topicID)
				if err != nil {
					log.Printf("  - error deleting merged topic %d: %v", topicID, err)
				}
			}
		}

		if !dryRun {
			// Rebuild stats for primary topic once per group
			err = cd.queries.SystemRebuildForumTopicMetaByID(ctx, primaryTopicID)
			if err != nil {
				log.Printf("  - error rebuilding meta for topic %d: %v", primaryTopicID, err)
			}
		}

		mergeGroups = append(mergeGroups, MergeGroup{
			PrimaryTopicID: primaryTopicID,
			MergedTopicIDs: mergedIDs,
			Participants:   strings.Split(key, ","),
		})
	}

	return mergeGroups, nil
}
