package common

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// MergePrivateTopicsWithSameParticipants finds private forum topics with identical participants
// and merges their threads into a single primary topic. The remaining topics are deleted.
// It returns the number of topics merged or an error.
func (cd *CoreData) MergePrivateTopicsWithSameParticipants(ctx context.Context, dryRun bool) (int, error) {
	topics, err := cd.Queries().AdminListAllPrivateTopics(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting private forum topics: %w", err)
	}

	topicParticipants := make(map[int32][]string)
	for _, topic := range topics {
		participants, err := cd.Queries().AdminListPrivateTopicParticipantsByTopicID(ctx, sql.NullInt32{Int32: topic.Idforumtopic, Valid: true})
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

	mergedCount := 0
	for key, topicIDs := range participantsToTopics {
		if len(topicIDs) <= 1 {
			continue
		}
		if key == "" {
			continue // Skip topics with no participants (should be handled by clean-empty)
		}

		sort.Slice(topicIDs, func(i, j int) bool {
			return topicIDs[i] < topicIDs[j]
		})

		primaryTopicID := topicIDs[0]
		log.Printf("Merging topics into primary topic ID %d (Participants: %s)", primaryTopicID, key)

		for _, topicID := range topicIDs[1:] {
			log.Printf("  - Merging from topic ID: %d", topicID)
			if !dryRun {
				// Move threads
				err := cd.Queries().AdminMoveThreadsToTopic(ctx, db.AdminMoveThreadsToTopicParams{
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
				grants, err := cd.Queries().AdminListGrantsByTopicID(ctx, sql.NullInt32{Int32: topicID, Valid: true})
				if err == nil {
					for _, grant := range grants {
						err = cd.Queries().AdminDeleteGrant(ctx, grant.ID)
						if err != nil {
							log.Printf("  - error deleting grant %d on topic %d: %v", grant.ID, topicID, err)
						}
					}
				}

				// Delete old topic
				err = cd.Queries().AdminDeleteForumTopic(ctx, topicID)
				if err != nil {
					log.Printf("  - error deleting merged topic %d: %v", topicID, err)
				}
			}
			mergedCount++
		}

		if !dryRun {
			// Rebuild stats for primary topic once per group
			err = cd.Queries().SystemRebuildForumTopicMetaByID(ctx, primaryTopicID)
			if err != nil {
				log.Printf("  - error rebuilding meta for topic %d: %v", primaryTopicID, err)
			}
		}
	}

	log.Printf("Merged %d topics.", mergedCount)
	return mergedCount, nil
}
