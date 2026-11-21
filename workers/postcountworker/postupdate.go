package postcountworker

import (
	"context"

	"github.com/arran4/goa4web/internal/postcount"
)

// PostUpdateQuerier is retained as a compatibility alias for the postcount contract.
type PostUpdateQuerier = postcount.PostUpdateQuerier

// PostUpdate refreshes metadata on the given forum thread and topic.
// It recalculates thread counters and rebuilds the topic aggregates.
func PostUpdate(ctx context.Context, q PostUpdateQuerier, threadID, topicID int32) error {
	return postcount.ApplyPostUpdate(ctx, q, threadID, topicID)
}
