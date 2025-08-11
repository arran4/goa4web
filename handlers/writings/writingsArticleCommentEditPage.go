package writings

import (
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/workers/postcountworker"
)

// ArticleCommentEditActionPage updates a comment on a writing and refreshes thread metadata.
func ArticleCommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	comment, err := cd.ArticleComment(r)
	if err != nil || comment == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if _, ok := core.GetSessionOrFail(w, r); !ok {
		return
	}

	thread, err := cd.UpdateWritingReply(comment.Idcomments, int32(languageId), text)
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: comment.Idcomments, ThreadID: thread.ID, TopicID: thread.TopicID}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting), http.StatusTemporaryRedirect)
}

// ArticleCommentEditActionCancelPage aborts editing a comment.
func ArticleCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.Article()
	if err != nil || writing == nil {
		http.Redirect(w, r, "/writings", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting), http.StatusTemporaryRedirect)
}
