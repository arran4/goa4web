package forumcommon

import (
	"database/sql"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

// RequireThreadAndTopic loads the thread specified in the URL and caches the
// associated topic on CoreData. Both records are accessible through CoreData
// for subsequent handlers.
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

		session, _ := core.GetSession(r)
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
			if cd.Config.LogFlags&config.LogFlagAccess != 0 {
				log.Printf("RequireThreadAndTopic thread uid=%d topic=%d thread=%d: %v", uid, topicID, threadID, err)
			}
			http.NotFound(w, r)
			return
		}

		topicRow, err := cd.ForumTopicByID(threadRow.ForumtopicIdforumtopic)
		if err != nil {
			if cd.Config.LogFlags&config.LogFlagAccess != 0 {
				log.Printf("RequireThreadAndTopic topic uid=%d topic=%d thread=%d: %v", uid, topicID, threadID, err)
			}
			http.NotFound(w, r)
			return
		}

		if int(topicRow.Idforumtopic) != topicID {
			if cd.Config.LogFlags&config.LogFlagAccess != 0 {
				log.Printf("RequireThreadAndTopic mismatch uid=%d urlTopic=%d threadTopic=%d", uid, topicID, topicRow.Idforumtopic)
			}
			http.NotFound(w, r)
			return
		}
		cd.SetCurrentThreadAndTopic(int32(threadID), threadRow.ForumtopicIdforumtopic)
		next.ServeHTTP(w, r)
	})
}
