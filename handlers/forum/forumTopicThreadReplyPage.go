package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	search "github.com/arran4/goa4web/handlers/search"
	"log"
	"net/http"
	"strconv"
)

func TopicThreadReplyPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	threadRow := r.Context().Value(hcommon.KeyThread).(*GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow)
	topicRow := r.Context().Value(hcommon.KeyTopic).(*GetForumTopicByIdForUserRow)

	queries := r.Context().Value(hcommon.KeyQueries).(*Queries)

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	provider := getEmailProvider()

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: threadRow.Idforumthread,
		Idusers:                  uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := notifyChange(r.Context(), provider, row.Username.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), ListUsersSubscribedToThreadParams{
		Idusers:                  uid,
		ForumthreadIdforumthread: threadRow.Idforumthread,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := notifyChange(r.Context(), provider, row.Username.String, endUrl); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	cid, err := queries.CreateComment(r.Context(), CreateCommentParams{
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

	wordIds, done := search.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if search.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	notifyThreadSubscribers(r.Context(), provider, queries, threadRow.Idforumthread, uid, endUrl)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	threadRow := r.Context().Value(hcommon.KeyThread).(*GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsRow)
	topicRow := r.Context().Value(hcommon.KeyTopic).(*GetForumTopicByIdForUserRow)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
