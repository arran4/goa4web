package writings

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/workers/postcountworker"
)

// ArticleCommentEditActionPage updates a comment on a writing and refreshes thread metadata.
func ArticleCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}
	text := r.PostFormValue("replytext")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}
	comment, err := cd.ArticleComment(r)
	if err != nil || comment == nil {
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}

	thread, err := cd.UpdateWritingReply(comment.Idcomments, int32(languageId), text)
	if err != nil {
		handlers.RedirectToGet(w, r, "?error="+err.Error())
		return
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: comment.Idcomments, ThreadID: thread.Idforumthread, TopicID: thread.ForumtopicIdforumtopic}
		}
	}

	handlers.RedirectToGet(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting))
}

// ArticleCommentEditActionCancelPage aborts editing a comment.
func ArticleCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		handlers.RedirectToGet(w, r, "/writings")
		return
	}
	handlers.RedirectToGet(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting))
}
