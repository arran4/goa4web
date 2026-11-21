package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// DeliverForumReplyJob loads current display data for an immutable reply activity.
// Deleted replies or resources have no notification to deliver.
func (n *Notifier) DeliverForumReplyJob(ctx context.Context, job *db.DurableJob) error {
	commentID, appendID := common.ActivityPositionCommentAndAppend(job.TargetRevision)
	key := fmt.Sprintf("comment:%d", commentID)
	if appendID != 0 {
		key = fmt.Sprintf("%s:append:%d", key, appendID)
	}
	if job.TargetType != "comment" || commentID <= 0 || appendID > math.MaxInt32 ||
		job.TargetID != int64(commentID) || job.TargetKey != key || !job.CreatedByUserID.Valid || job.CreatedByUserID.Int32 <= 0 {
		return fmt.Errorf("invalid forum reply identity for job %d", job.ID)
	}

	actorID := job.CreatedByUserID.Int32
	source, err := n.Queries.SystemGetForumReplyNotificationSource(ctx, db.SystemGetForumReplyNotificationSourceParams{
		CommentID: commentID, ActorID: actorID, AppendID: int32(appendID),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load forum reply notification source: %w", err)
	}
	private := source.Handler == "private"
	prefix := "Forum"
	if private {
		prefix = "Private Forum"
	}
	intent := common.ForumReplyNotificationIntent{
		ActorID:          actorID,
		Username:         source.Username.String,
		TopicID:          source.TopicID,
		TopicTitle:       source.TopicTitle.String,
		ThreadID:         source.ThreadID,
		CommentID:        commentID,
		AppendID:         appendID,
		Appended:         appendID != 0,
		ActivityID:       key,
		ActivityPosition: job.TargetRevision,
		Private:          private,
		CommentAnchor:    fmt.Sprintf("comment-%d", commentID),
		CommentText:      source.ReplyText.String,
		SubjectPrefix:    prefix,
	}
	return n.notifyForumReply(ctx, nil, intent, nil, true)
}
