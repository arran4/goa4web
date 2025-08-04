package forum

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/gorilla/mux"
)

func TopicThreadCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Edit Comment"
	queries := cd.Queries()
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	commentId, _ := strconv.Atoi(mux.Vars(r)["comment"])

	err = queries.UpdateCommentForCommenter(r.Context(), db.UpdateCommentForCommenterParams{
		CommentID:      int32(commentId),
		GrantCommentID: sql.NullInt32{Int32: int32(commentId), Valid: true},
		LanguageID:     int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
		GranteeID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		CommenterID: cd.UserID,
	})
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: threadRow.Idforumthread, TopicID: topicRow.Idforumtopic}
			evt.Data["CommentURL"] = cd.AbsoluteURL(fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicRow.Idforumtopic, threadRow.Idforumthread, commentId))
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicRow.Idforumtopic, threadRow.Idforumthread, commentId), http.StatusTemporaryRedirect)
}

func TopicThreadCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Forum - Edit Comment"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusTemporaryRedirect)
}
