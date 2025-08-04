package writings

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core"
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
	writing, err := cd.CurrentWriting()
	if err != nil || writing == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	comment, err := cd.CurrentComment(r)
	if err != nil || comment == nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

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

	uid = cd.UserID
	if err = queries.UpdateCommentForCommenter(r.Context(), db.UpdateCommentForCommenterParams{
		CommentID:      comment.Idcomments,
		GrantCommentID: sql.NullInt32{Int32: comment.Idcomments, Valid: true},
		LanguageID:     int32(languageId),
		Text:           sql.NullString{String: text, Valid: true},
		GranteeID:      sql.NullInt32{Int32: uid, Valid: uid != 0},
		CommenterID:    uid,
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

	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting), http.StatusTemporaryRedirect)
}

// ArticleCommentEditActionCancelPage aborts editing a comment.
func ArticleCommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	writing, err := cd.CurrentWriting()
	if err != nil || writing == nil {
		http.Redirect(w, r, "/writings", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/writings/article/%d", writing.Idwriting), http.StatusTemporaryRedirect)
}
