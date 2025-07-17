package searchworker

import (
	"context"
	"fmt"

	db "github.com/arran4/goa4web/internal/db"
)

// PostUpdate refreshes metadata on the given forum thread and topic.
// It recalculates thread counters and rebuilds the topic aggregates.
//
// The helper replaces an older one-shot SQL statement:
//
//	UPDATE comments c, forumthread th, forumtopic t
//	SET
//	    th.lastposter=c.users_idusers, t.lastposter=c.users_idusers,
//	    th.lastaddition=c.written, t.lastaddition=c.written,
//	    t.comments=IF(th.comments IS NULL, 0, t.comments+1),
//	    t.threads=IF(th.comments IS NULL, IF(t.threads IS NULL, 1, t.threads+1), t.threads),
//	    th.comments=IF(th.comments IS NULL, 0, th.comments+1),
//	    th.firstpost=IF(th.firstpost=0, c.idcomments, th.firstpost)
//	WHERE c.idcomments=?;
//
// The same effect is achieved by calling RecalculateForumThreadByIdMetaData
// and RebuildForumTopicByIdMetaColumns generated via sqlc.
func PostUpdate(ctx context.Context, q *db.Queries, threadID, topicID int32) error {
	if err := q.RecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.RebuildForumTopicByIdMetaColumns(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}
