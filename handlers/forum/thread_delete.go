package forum

import (
	"context"
	"github.com/arran4/goa4web/internal/db"
)

// ThreadDelete removes a forum thread and updates topic statistics.
func ThreadDelete(ctx context.Context, q *db.Queries, threadID, topicID int32) error {
	if err := q.AdminDeleteForumThread(ctx, threadID); err != nil {
		return err
	}
	return q.RebuildForumTopicByIdMetaColumns(ctx, topicID)
}
