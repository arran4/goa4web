package news

import (
	"context"
	"fmt"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func PostUpdateLocal(ctx context.Context, q *db.Queries, threadID, topicID int32) error {
	if err := q.RecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.RebuildForumTopicByIdMetaColumns(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}

// canEditNewsPost reports whether cd can modify the specified news post.
func canEditNewsPost(cd *hcommon.CoreData, postID int32) bool {
	if cd == nil {
		return false
	}
	if cd.HasGrant("news", "post", "edit", postID) && (cd.AdminMode || cd.UserID != 0) {
		return true
	}
	return false
}
