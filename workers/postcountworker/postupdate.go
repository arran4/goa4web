package postcountworker

import (
	"context"
	"fmt"
)

// PostUpdateQuerier defines the database operations needed to refresh post metadata.
type PostUpdateQuerier interface {
	AdminRecalculateForumThreadByIdMetaData(ctx context.Context, idforumthread int32) error
	SystemRebuildForumTopicMetaByID(ctx context.Context, idforumtopic int32) error
}

// PostUpdate refreshes metadata on the given forum thread and topic.
// It recalculates thread counters and rebuilds the topic aggregates.
func PostUpdate(ctx context.Context, q PostUpdateQuerier, threadID, topicID int32) error {
	if err := q.AdminRecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.SystemRebuildForumTopicMetaByID(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}
