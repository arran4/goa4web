package linker

import (
	"context"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"
)

// PostUpdate refreshes thread and topic metadata.
func PostUpdate(ctx context.Context, q *db.Queries, threadID, topicID int32) error {
	if err := q.RecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.RebuildForumTopicByIdMetaColumns(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}
