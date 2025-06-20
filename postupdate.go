package main

import "context"

// PostUpdate refreshes metadata on the given forum thread and topic.
// It recalculates thread counters and rebuilds the topic aggregates.
func PostUpdate(ctx context.Context, q *Queries, threadID, topicID int32) error {
	if err := q.RecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return err
	}
	if err := q.RebuildForumTopicByIdMetaColumns(ctx, topicID); err != nil {
		return err
	}
	return nil
}
