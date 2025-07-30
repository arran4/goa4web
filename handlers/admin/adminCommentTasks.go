package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

// DeleteCommentTask permanently removes a comment.
type DeleteCommentTask struct{ tasks.TaskString }

var deleteCommentTask = &DeleteCommentTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteCommentTask)(nil)

func (DeleteCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	q := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := q.ScrubComment(r.Context(), db.ScrubCommentParams{Text: sql.NullString{String: "", Valid: true}, Idcomments: int32(id)}); err != nil {
		return fmt.Errorf("delete comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

// EditCommentTask updates a comment's text.
type EditCommentTask struct{ tasks.TaskString }

var editCommentTask = &EditCommentTask{TaskString: TaskEdit}

var _ tasks.Task = (*EditCommentTask)(nil)

func (EditCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	text := r.PostFormValue("replytext")
	q := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := q.UpdateComment(r.Context(), db.UpdateCommentParams{Idcomments: int32(id), Text: sql.NullString{String: text, Valid: true}, LanguageIdlanguage: 0}); err != nil {
		return fmt.Errorf("edit comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

// BanCommentTask archives and scrubs a comment.
type BanCommentTask struct{ tasks.TaskString }

var banCommentTask = &BanCommentTask{TaskString: "Ban"}

var _ tasks.Task = (*BanCommentTask)(nil)

func (BanCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	q := cd.Queries()
	c, err := q.GetCommentById(r.Context(), int32(id))
	if err != nil {
		return fmt.Errorf("fetch comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := q.ArchiveComment(r.Context(), db.ArchiveCommentParams{
		Idcomments:         c.Idcomments,
		ForumthreadID:      c.ForumthreadID,
		UsersIdusers:       c.UsersIdusers,
		LanguageIdlanguage: c.LanguageIdlanguage,
		Written:            c.Written,
		Text:               c.Text,
	}); err != nil {
		return fmt.Errorf("archive comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := q.ScrubComment(r.Context(), db.ScrubCommentParams{Text: sql.NullString{String: "", Valid: true}, Idcomments: c.Idcomments}); err != nil {
		return fmt.Errorf("scrub comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
