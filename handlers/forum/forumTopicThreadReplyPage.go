package forum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"

	"github.com/arran4/goa4web/internal/emailutil"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	searchutil "github.com/arran4/goa4web/internal/searchutil"
)

func TopicThreadReplyPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	threadRow := r.Context().Value(hcommon.KeyThread).(*db.GetThreadByIdForUserByIdWithLastPosterUserNameAndPermissionsRow)
	topicRow := r.Context().Value(hcommon.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	if evt, ok := r.Context().Value(hcommon.KeyBusEvent).(*eventbus.Event); ok && evt != nil {
		evt.Item = notif.ForumReplyInfo{TopicTitle: topicRow.Title.String, ThreadID: threadRow.Idforumthread, Thread: threadRow}
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	provider := email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig)

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), db.ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: threadRow.Idforumthread,
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := emailutil.NotifyChange(r.Context(), provider, row.Username.String, endUrl, "update", nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), db.ListUsersSubscribedToThreadParams{
		Idusers:                  uid,
		ForumthreadIdforumthread: threadRow.Idforumthread,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := emailutil.NotifyChange(r.Context(), provider, row.Username.String, endUrl, "update", nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage:       int32(languageId),
		UsersIdusers:             uid,
		ForumthreadIdforumthread: threadRow.Idforumthread,
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	})
	if err != nil {
		log.Printf("Error: CreateComment: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := PostUpdate(r.Context(), queries, threadRow.Idforumthread, topicRow.Idforumtopic); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchutil.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	notifyThreadSubscribers(r.Context(), provider, queries, threadRow.Idforumthread, uid, endUrl)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	threadRow := r.Context().Value(hcommon.KeyThread).(*db.GetThreadByIdForUserByIdWithLastPosterUserNameAndPermissionsRow)
	topicRow := r.Context().Value(hcommon.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
