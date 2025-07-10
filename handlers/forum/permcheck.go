package forum

import (
	"net/http"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// userTopicLevel returns the user's level for the given topic.
func userTopicLevel(r *http.Request, q *db.Queries, tid int32) int32 {
	session, _ := core.GetSession(r)
	uid, _ := session.Values["UID"].(int32)
	if uid == 0 {
		return 0
	}
	row, err := q.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), db.GetUsersTopicLevelByUserIdAndThreadIdParams{
		UsersIdusers:           uid,
		ForumtopicIdforumtopic: tid,
	})
	if err != nil || !row.Level.Valid {
		return 0
	}
	return row.Level.Int32
}

// topicRestriction fetches restriction details for the topic.
func topicRestriction(r *http.Request, q *db.Queries, tid int32) *db.GetForumTopicRestrictionsByForumTopicIdRow {
	rows, err := q.GetForumTopicRestrictionsByForumTopicId(r.Context(), tid)
	if err != nil || len(rows) == 0 {
		return nil
	}
	return rows[0]
}

// CanReply reports whether the current user may reply in the given topic.
func CanReply(r *http.Request, tid int32) bool {
	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	if !cd.HasRole("writer") {
		return false
	}
	q := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	res := topicRestriction(r, q, tid)
	if res == nil || !res.Replylevel.Valid {
		return true
	}
	return userTopicLevel(r, q, tid) >= res.Replylevel.Int32
}

// CanCreateThread reports whether the current user may create a new thread in the topic.
func CanCreateThread(r *http.Request, tid int32) bool {
	cd := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	if !cd.HasRole("writer") {
		return false
	}
	q := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	res := topicRestriction(r, q, tid)
	if res == nil || !res.Newthreadlevel.Valid {
		return true
	}
	return userTopicLevel(r, q, tid) >= res.Newthreadlevel.Int32
}
