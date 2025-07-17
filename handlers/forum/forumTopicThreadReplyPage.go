package forum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"

	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	searchworker "github.com/arran4/goa4web/internal/searchworker"
)

func TopicThreadReplyPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	threadRow := r.Context().Value(hcommon.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(hcommon.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	if cd, ok := r.Context().Value(hcommon.KeyCoreData).(*hcommon.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["reply"] = notif.ForumReplyInfo{TopicTitle: topicRow.Title.String, ThreadID: threadRow.Idforumthread, Thread: threadRow}
		}
	}

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	provider := email.ProviderFromConfig(config.AppRuntimeConfig)

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), db.ListUsersSubscribedToThreadParams{
		ForumthreadID: threadRow.Idforumthread,
		Idusers:       uid,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := notif.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, endUrl, "update", nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)
			}
		}
	}

	if rows, err := queries.ListUsersSubscribedToThread(r.Context(), db.ListUsersSubscribedToThreadParams{
		Idusers:       uid,
		ForumthreadID: threadRow.Idforumthread,
	}); err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
	} else if provider != nil {
		for _, row := range rows {
			if err := notif.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, endUrl, "update", nil); err != nil {
				log.Printf("Error: notifyChange: %s", err)

			}
		}
	}

	cid, err := queries.CreateComment(r.Context(), db.CreateCommentParams{
		LanguageIdlanguage: int32(languageId),
		UsersIdusers:       uid,
		ForumthreadID:      threadRow.Idforumthread,
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

	// TODO move to searchworker that is automatically activated by a event.
	wordIds, done := searchworker.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchworker.InsertWordsToForumSearch(w, r, wordIds, queries, cid) {
		return
	}

	// TODO remove and replace with proper eventbus notification
	notif.Notifier{EmailProvider: provider, Queries: queries}.NotifyThreadSubscribers(r.Context(), threadRow.Idforumthread, uid, endUrl)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	threadRow := r.Context().Value(hcommon.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(hcommon.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
