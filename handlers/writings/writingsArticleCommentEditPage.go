package writings

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
)

// ArticleCommentEditActionPage updates a comment on a writing and refreshes thread metadata.
func ArticleCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		handlers.RedirectSeeOtherWithMessage(w, r, "", err.Error())
		return
	}
	text := r.PostFormValue("replytext")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		handlers.RedirectSeeOtherWithMessage(w, r, "", err.Error())
		return
	}
	comment, err := cd.ArticleComment(r)
	if err != nil || comment == nil {
		handlers.RedirectSeeOtherWithMessage(w, r, "", err.Error())
		return
	}

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}

	thread, err := cd.UpdateWritingReply(comment.Idcomments, int32(languageId), text)
	if err != nil {
		handlers.RedirectSeeOtherWithMessage(w, r, "", err.Error())
		return
	}
	if err := cd.HandleThreadUpdated(r.Context(), common.ThreadUpdatedEvent{
		ThreadID:             thread.Idforumthread,
		TopicID:              thread.ForumtopicIdforumtopic,
		CommentID:            comment.Idcomments,
		LabelItem:            "writing",
		LabelItemID:          writing.Idwriting,
		ClearUnreadForOthers: true,
		MarkThreadRead:       true,
		IncludePostCount:     true,
	}); err != nil {
		log.Printf("writing comment edit side effects: %v", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting), http.StatusSeeOther)
}

// ArticleCommentEditActionCancelPage aborts editing a comment.
func ArticleCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		http.Redirect(w, r, "/writings", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting), http.StatusSeeOther)
}
