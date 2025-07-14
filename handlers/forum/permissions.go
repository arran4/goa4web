package forum

import (
	"context"
	"database/sql"

	db "github.com/arran4/goa4web/internal/db"
)

// UserCanCreateThread reports whether uid may create a thread in the topic.
func UserCanCreateThread(ctx context.Context, q *db.Queries, topicID, uid int32) (bool, error) {
	rows, err := q.GetForumTopicRestrictionsByForumTopicId(ctx, topicID)
	if err != nil {
		return false, err
	}
	var required int32
	if len(rows) > 0 && rows[0].NewthreadRoleID.Valid {
		required = rows[0].NewthreadRoleID.Int32
	}

	level, err := q.GetUsersTopicLevelByUserIdAndThreadId(ctx, db.GetUsersTopicLevelByUserIdAndThreadIdParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: topicID,
	})
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	var have int32
	if err == nil && level.RoleID.Valid {
		have = level.RoleID.Int32
	}

	return have >= required, nil
}
