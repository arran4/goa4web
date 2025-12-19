package forum

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/gorilla/mux"
)

func TopicThreadCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	text := r.PostFormValue("replytext")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Forum - Edit Comment"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	commentId, _ := strconv.Atoi(mux.Vars(r)["comment"])

	err = cd.UpdateForumComment(int32(commentId), int32(languageId), text)
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}

	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:         threadRow.Idforumthread,
		TopicID:          topicRow.Idforumtopic,
		CommentID:        int32(commentId),
		CommentURL:       cd.AbsoluteURL(fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicRow.Idforumtopic, threadRow.Idforumthread, commentId)),
		IncludePostCount: true,
	}); err != nil {
		log.Printf("thread comment edit side effects: %v", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/forum/topic/%d/thread/%d#comment-%d", topicRow.Idforumtopic, threadRow.Idforumthread, commentId), http.StatusSeeOther)
}

func TopicThreadCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	cd.PageTitle = "Forum - Edit Comment"
	threadRow, err := cd.SelectedThread()
	if err != nil || threadRow == nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}
	topicRow, err := cd.CurrentTopic()
	if err != nil || topicRow == nil {
		handlers.RedirectSeeOtherWithError(w, r, "", err)
		return
	}

	endUrl := fmt.Sprintf("/forum/topic/%d/thread/%d#bottom", topicRow.Idforumtopic, threadRow.Idforumthread)

	http.Redirect(w, r, endUrl, http.StatusSeeOther)
}
