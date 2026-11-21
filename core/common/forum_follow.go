package common

import (
	"fmt"

	"github.com/arran4/goa4web/internal/db"
)

// PublicForumNewThreadsPattern is the explicit forum-wide in-site subscription for newly created public threads.
const PublicForumNewThreadsPattern = "create thread:/forum/*"

// FollowsPublicForumNewThreads reports the current user's direct forum follow.
func (cd *CoreData) FollowsPublicForumNewThreads() bool {
	return cd.Subscribed(PublicForumNewThreadsPattern, "internal")
}

// SetPublicForumNewThreadFollow updates only the signed-in user's public forum new-thread scope.
func (cd *CoreData) SetPublicForumNewThreadFollow(follow bool) error {
	if cd.UserID <= 0 || cd.queries == nil {
		return fmt.Errorf("set forum follow: signed-in user required")
	}
	if follow {
		return EnsureSubscriptionIdempotent(cd.ctx, cd.queries, cd.UserID, PublicForumNewThreadsPattern, "internal")
	}
	return cd.queries.DeleteSubscriptionForSubscriber(cd.ctx, db.DeleteSubscriptionForSubscriberParams{
		SubscriberID: cd.UserID,
		Pattern:      PublicForumNewThreadsPattern,
		Method:       "internal",
	})
}
