// Package forum owns forum-domain event contracts and operations.
package forum

import (
	"context"
	"fmt"
)

// PostCountEventKey identifies metadata refresh requests in an event payload.
const PostCountEventKey = "post_count"

// PostCountUpdate describes which comment, thread, and topic counts to refresh.
type PostCountUpdate struct {
	ThreadID  int32
	TopicID   int32
	CommentID int32
}

// PostUpdateQuerier defines the database operations needed to refresh forum metadata.
type PostUpdateQuerier interface {
	AdminRecalculateForumThreadByIdMetaData(ctx context.Context, idforumthread int32) error
	SystemRebuildForumTopicMetaByID(ctx context.Context, idforumtopic int32) error
}

// ApplyPostUpdate refreshes metadata on the given forum thread and topic.
func ApplyPostUpdate(ctx context.Context, q PostUpdateQuerier, threadID, topicID int32) error {
	if err := q.AdminRecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.SystemRebuildForumTopicMetaByID(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}
