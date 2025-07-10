package forum

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/arran4/goa4web/internal/db"
)

// canCreateThread returns true if user uid is allowed to create
// a new thread in the given topic according to topic restrictions.
func canCreateThread(ctx context.Context, q *db.Queries, topicID, uid int32) (bool, error) {
	restr, err := q.GetForumTopicRestrictionsByForumTopicId(ctx, topicID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	var required int32
	if len(restr) > 0 && restr[0].Newthreadlevel.Valid {
		required = restr[0].Newthreadlevel.Int32
	}
	// No restriction -> allowed.
	if required == 0 {
		return true, nil
	}

	userLevel, err := q.GetUsersTopicLevelByUserIdAndThreadId(ctx, db.GetUsersTopicLevelByUserIdAndThreadIdParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: topicID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	var level int32
	if userLevel != nil && userLevel.Level.Valid {
		level = userLevel.Level.Int32
	}
	return level >= required, nil
}
