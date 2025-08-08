package admin

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// DeleteCommentTask permanently removes a comment.
type DeleteCommentTask struct{ tasks.TaskString }

var deleteCommentTask = &DeleteCommentTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteCommentTask)(nil)

func (DeleteCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return fmt.Errorf("delete comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.Queries().AdminScrubComment(r.Context(), db.AdminScrubCommentParams{Text: sql.NullString{String: "", Valid: true}, Idcomments: c.Idcomments}); err != nil {
		return fmt.Errorf("delete comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

// EditCommentTask updates a comment's text.
type EditCommentTask struct{ tasks.TaskString }

var editCommentTask = &EditCommentTask{TaskString: TaskEdit}

var _ tasks.Task = (*EditCommentTask)(nil)

func (EditCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return fmt.Errorf("edit comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	text := r.PostFormValue("replytext")
	if err := cd.Queries().AdminScrubComment(r.Context(), db.AdminScrubCommentParams{Text: sql.NullString{String: text, Valid: true}, Idcomments: c.Idcomments}); err != nil {
		return fmt.Errorf("edit comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

// BanCommentTask archives and scrubs a comment.
type BanCommentTask struct{ tasks.TaskString }

var banCommentTask = &BanCommentTask{TaskString: "Ban"}

var _ tasks.Task = (*BanCommentTask)(nil)

func (BanCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return fmt.Errorf("fetch comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	q := cd.Queries()
	if err := q.AdminArchiveComment(r.Context(), db.AdminArchiveCommentParams{
		Idcomments:         c.Idcomments,
		ForumthreadID:      c.ForumthreadID,
		UsersIdusers:       c.UsersIdusers,
		LanguageIdlanguage: c.LanguageIdlanguage.Int32,
		Written:            c.Written,
		Text:               c.Text,
	}); err != nil {
		return fmt.Errorf("archive comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := q.AdminScrubComment(r.Context(), db.AdminScrubCommentParams{Text: sql.NullString{String: "", Valid: true}, Idcomments: c.Idcomments}); err != nil {
		return fmt.Errorf("scrub comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
