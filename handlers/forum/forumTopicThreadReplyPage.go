package forum

import (
	"database/sql"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/config"

	"github.com/arran4/goa4web/core"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
	searchworker "github.com/arran4/goa4web/internal/searchworker"

	"github.com/arran4/goa4web/internal/tasks"
)

// ReplyTask handles replying to an existing thread.
type ReplyTask struct{ tasks.TaskString }

var replyTask = &ReplyTask{TaskString: TaskReply}

func (ReplyTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}

	threadRow := r.Context().Value(handlers.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(handlers.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	if cd, ok := r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["reply"] = notif.ForumReplyInfo{TopicTitle: topicRow.Title.String, ThreadID: threadRow.Idforumthread, Thread: threadRow}
		}
	}

	queries := r.Context().Value(handlers.KeyQueries).(*db.Queries)

	text := r.PostFormValue("replytext")
	languageId, _ := strconv.Atoi(r.PostFormValue("language"))
	uid, _ := session.Values["UID"].(int32)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

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
	if cd, ok := r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: cid, Text: text}
		}
	}

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}

func TopicThreadReplyCancelPage(w http.ResponseWriter, r *http.Request) {
	threadRow := r.Context().Value(handlers.KeyThread).(*db.GetThreadLastPosterAndPermsRow)
	topicRow := r.Context().Value(handlers.KeyTopic).(*db.GetForumTopicByIdForUserRow)

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
