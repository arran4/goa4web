package linker

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core"
	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

// CommentEditActionPage updates a comment then refreshes thread metadata.
func CommentEditActionPage(w http.ResponseWriter, r *http.Request) {
	languageId, err := strconv.Atoi(r.PostFormValue("language"))
	if err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	text := r.PostFormValue("replytext")

	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	commentId, _ := strconv.Atoi(vars["comment"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	comment := r.Context().Value(hcommon.KeyComment).(*db.GetCommentByIdForUserRow)

	thread, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		UsersIdusers:  uid,
		Idforumthread: comment.ForumthreadID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("Error: getThreadByIdForUserByIdWithLastPosterUserNameAndPermissions: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return
		}
	}

	if err = queries.UpdateComment(r.Context(), db.UpdateCommentParams{
		Idcomments:         int32(commentId),
		LanguageIdlanguage: int32(languageId),
		Text: sql.NullString{
			String: text,
			Valid:  true,
		},
	}); err != nil {
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	if err := PostUpdate(r.Context(), queries, thread.Idforumthread, thread.ForumtopicIdforumtopic); err != nil {
		log.Printf("Error: postUpdate: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/linker/comments/%d", linkId), http.StatusTemporaryRedirect)
}

// CommentEditActionCancelPage aborts editing a comment.
func CommentEditActionCancelPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	linkId, _ := strconv.Atoi(vars["link"])
	http.Redirect(w, r, fmt.Sprintf("/linker/comments/%d", linkId), http.StatusTemporaryRedirect)
}
