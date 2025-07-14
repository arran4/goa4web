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

// canEditNewsPost returns true if cd has permission to edit a news post by authorID.
func canEditNewsPost(cd *hcommon.CoreData, authorID int32) bool {
	if cd == nil {
		return false
	}
	if cd.HasRole("administrator") && cd.AdminMode {
		return true
	}
	return cd.HasRole("content writer") && cd.UserID == authorID
}
