package postcountworker

import (
	"context"
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// PostUpdate refreshes metadata on the given forum thread and topic.
// It recalculates thread counters and rebuilds the topic aggregates.
func PostUpdate(ctx context.Context, q db.Querier, threadID, topicID int32) error {
	if err := q.AdminRecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.SystemRebuildForumTopicMetaByID(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}
