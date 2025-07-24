package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
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
	queries := cd.Queries()
	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])
	commentId, _ := strconv.Atoi(vars["comment"])

	session, ok := cd.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	comment := r.Context().Value(consts.KeyComment).(*db.GetCommentByIdForUserRow)

	thread, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      comment.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error: getThreadLastPosterAndPerms: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err = queries.UpdateComment(r.Context(), db.UpdateCommentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text:               sql.NullString{String: text, Valid: true},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{ThreadID: thread.Idforumthread, TopicID: thread.ForumtopicIdforumtopic}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", articleId), http.StatusTemporaryRedirect)
}

// ArticleCommentEditActionCancelPage aborts editing a comment.
func ArticleCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleId, _ := strconv.Atoi(vars["article"])
	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", articleId), http.StatusTemporaryRedirect)
}
