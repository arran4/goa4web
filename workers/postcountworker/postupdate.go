package postcountworker

import (
	"context"

	forumdomain "github.com/arran4/goa4web/internal/forum"
)

// PostUpdateQuerier is retained as a compatibility alias for the forum contract.
type PostUpdateQuerier = forumdomain.PostUpdateQuerier

// PostUpdate refreshes metadata on the given forum thread and topic.
// It recalculates thread counters and rebuilds the topic aggregates.
func PostUpdate(ctx context.Context, q PostUpdateQuerier, threadID, topicID int32) error {
	return forumdomain.ApplyPostUpdate(ctx, q, threadID, topicID)
}
