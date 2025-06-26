package forum

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

// GetThreadAndTopic retrieves thread and topic details for the request and stores them in the context.
func GetThreadAndTopic() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		topicID, err := strconv.Atoi(vars["topic"])
		if err != nil {
			return false
		}
		threadID, err := strconv.Atoi(vars["thread"])
		if err != nil {
			return false
		}

		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

		session, _ := core.GetSession(r)
		var uid int32
		if session != nil {
			uid, _ = session.Values["UID"].(int32)
		}

		threadRow, err := queries.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissions(r.Context(), db.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams{
			UsersIdusers:  uid,
			Idforumthread: int32(threadID),
		})
		if err != nil {
			if runtimeconfig.AppRuntimeConfig.LogFlags&runtimeconfig.LogFlagAccess != 0 {
				log.Printf("GetThreadAndTopic thread uid=%d topic=%d thread=%d: %v", uid, topicID, threadID, err)
			}
			return false
		}

		topicRow, err := queries.GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
			UsersIdusers: uid,
			Idforumtopic: threadRow.ForumtopicIdforumtopic,
		})
		if err != nil {
			if runtimeconfig.AppRuntimeConfig.LogFlags&runtimeconfig.LogFlagAccess != 0 {
				log.Printf("GetThreadAndTopic topic uid=%d topic=%d thread=%d: %v", uid, topicID, threadID, err)
			}
			return false
		}

		if int(topicRow.Idforumtopic) != topicID {
			if runtimeconfig.AppRuntimeConfig.LogFlags&runtimeconfig.LogFlagAccess != 0 {
				log.Printf("GetThreadAndTopic mismatch uid=%d urlTopic=%d threadTopic=%d", uid, topicID, topicRow.Idforumtopic)
			}
			return false
		}

		ctx := context.WithValue(r.Context(), hcommon.KeyThread, threadRow)
		ctx = context.WithValue(ctx, hcommon.KeyTopic, topicRow)
		*r = *r.WithContext(ctx)
		return true
	}
}

// TargetUsersLevelNotHigherThanAdminsMax verifies the target user's level does not exceed the admin's maximum.
func TargetUsersLevelNotHigherThanAdminsMax() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		adminUID, _ := session.Values["UID"].(int32)

		targetUID, err := strconv.Atoi(r.PostFormValue("uid"))
		if err != nil {
			return false
		}

		tid, err := strconv.Atoi(r.PostFormValue("tid"))
		if err != nil {
			return false
		}

		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

		targetUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), db.GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(targetUID),
		})
		if err != nil {
			return false
		}

		adminUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), db.GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(adminUID),
		})
		if err != nil {
			return false
		}

		return adminUser.Invitemax.Int32 >= targetUser.Level.Int32
	}
}

// AdminUsersMaxLevelNotLowerThanTargetLevel ensures the admin's max level exceeds the requested level values.
func AdminUsersMaxLevelNotLowerThanTargetLevel() mux.MatcherFunc {
	return func(r *http.Request, m *mux.RouteMatch) bool {
		session, err := core.GetSession(r)
		if err != nil {
			return false
		}
		adminUID, _ := session.Values["UID"].(int32)

		inviteMax, err := strconv.Atoi(r.PostFormValue("inviteMax"))
		if err != nil {
			return false
		}
		level, err := strconv.Atoi(r.PostFormValue("level"))
		if err != nil {
			return false
		}
		tid, err := strconv.Atoi(r.PostFormValue("tid"))
		if err != nil {
			return false
		}
		queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

		adminUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), db.GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(adminUID),
		})
		if err != nil {
			return false
		}

		return int(adminUser.Invitemax.Int32) >= level && int(adminUser.Invitemax.Int32) >= inviteMax
	}
}
