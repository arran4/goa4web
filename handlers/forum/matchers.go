package forum

import (
	"context"
	"database/sql"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

// RequireThreadAndTopic loads the thread and topic specified in the URL and
// verifies that they belong together before passing control to the next
// handler. The loaded rows are stored on the request context.
func RequireThreadAndTopic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		topicID, err := strconv.Atoi(vars["topic"])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		threadID, err := strconv.Atoi(vars["thread"])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		queries := cd.Queries()

		session, _ := cd.GetSession(r)
		var uid int32
		if session != nil {
			uid, _ = session.Values["UID"].(int32)
		}

		threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
			ViewerID:      uid,
			ThreadID:      int32(threadID),
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			if config.AppRuntimeConfig.LogFlags&config.LogFlagAccess != 0 {
				log.Printf("RequireThreadAndTopic thread uid=%d topic=%d thread=%d: %v", uid, topicID, threadID, err)
			}
			http.NotFound(w, r)
			return
		}

		topicRow, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
			ViewerID:      uid,
			Idforumtopic:  threadRow.ForumtopicIdforumtopic,
			ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		})
		if err != nil {
			if config.AppRuntimeConfig.LogFlags&config.LogFlagAccess != 0 {
				log.Printf("RequireThreadAndTopic topic uid=%d topic=%d thread=%d: %v", uid, topicID, threadID, err)
			}
			http.NotFound(w, r)
			return
		}

		if int(topicRow.Idforumtopic) != topicID {
			if config.AppRuntimeConfig.LogFlags&config.LogFlagAccess != 0 {
				log.Printf("RequireThreadAndTopic mismatch uid=%d urlTopic=%d threadTopic=%d", uid, topicID, topicRow.Idforumtopic)
			}
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), consts.KeyThread, threadRow)
		ctx = context.WithValue(ctx, consts.KeyTopic, topicRow)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
