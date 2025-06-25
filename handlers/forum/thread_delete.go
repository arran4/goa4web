package forum

import "context"

// ThreadDelete removes a forum thread and updates topic statistics.
func ThreadDelete(ctx context.Context, q *Queries, threadID, topicID int32) error {
	if err := q.DeleteForumThread(ctx, threadID); err != nil {
		return err
	}
	return q.RebuildForumTopicByIdMetaColumns(ctx, topicID)
}
