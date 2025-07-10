package blogs

import (
	"context"
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// userMayReply reports whether uid may reply to blog post b.
func userMayReply(ctx context.Context, q *db.Queries, b *db.GetBlogEntryForUserByIdRow, uid int32) bool {
	if uid == 0 {
		return false
	}
	if b.ForumthreadID == 0 {
		return true
	}
	th, err := q.GetThreadLastPosterAndPerms(ctx, db.GetThreadLastPosterAndPermsParams{
		UsersIdusers:  uid,
		Idforumthread: b.ForumthreadID,
	})
	if err != nil {
		return false
	}
	return !(th.Locked.Valid && th.Locked.Bool)
}

// currentUserMayReply returns whether the current request user may reply to b.
func currentUserMayReply(r *http.Request, b *db.GetBlogEntryForUserByIdRow) bool {
	cd, _ := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData)
	q, _ := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	if q == nil || cd == nil {
		return false
	}
	return userMayReply(r.Context(), q, b, cd.UserID)
}
