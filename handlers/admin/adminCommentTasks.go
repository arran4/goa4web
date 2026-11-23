package admin

import (
	"fmt"
	"math"
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
	if err := cd.AdminScrubComment(r.Context(), c.Idcomments, "", false); err != nil {
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
	if err := cd.AdminEditComment(r.Context(), c.Idcomments, text); err != nil {
		return fmt.Errorf("edit comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

// DeactivateCommentTask archives and scrubs a comment if not already deactivated.
type DeactivateCommentTask struct{ tasks.TaskString }

var deactivateCommentTask = &DeactivateCommentTask{TaskString: TaskDeactivate}

var _ tasks.Task = (*DeactivateCommentTask)(nil)

func (DeactivateCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return fmt.Errorf("fetch comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	q := cd.Queries()
	deactivated, err := q.AdminIsCommentDeactivated(r.Context(), c.Idcomments)
	if err != nil {
		return fmt.Errorf("check deactivated %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if deactivated {
		return fmt.Errorf("comment already deactivated %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("already deactivated")))
	}
	if err := cd.AdminDeactivateComment(r.Context(), db.AdminArchiveCommentParams{
		Idcomments:    c.Idcomments,
		ForumthreadID: c.ForumthreadID,
		UsersIdusers:  c.UsersIdusers,
		LanguageID:    c.LanguageID,
		Written:       c.Written,
		Text:          c.Text,
		Timezone:      c.Timezone,
	}, ""); err != nil {
		return fmt.Errorf("deactivate comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}

// RestoreCommentTask restores a previously deactivated comment.
type RestoreCommentTask struct{ tasks.TaskString }

var restoreCommentTask = &RestoreCommentTask{TaskString: TaskActivate}

var _ tasks.Task = (*RestoreCommentTask)(nil)

func (RestoreCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return fmt.Errorf("fetch comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	q := cd.Queries()
	deactivated, err := q.AdminIsCommentDeactivated(r.Context(), c.Idcomments)
	if err != nil {
		return fmt.Errorf("check deactivated %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !deactivated {
		return fmt.Errorf("comment not deactivated %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("not deactivated")))
	}
	rows, err := q.AdminListDeactivatedComments(r.Context(), db.AdminListDeactivatedCommentsParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return fmt.Errorf("list deactivated %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	var found *db.AdminListDeactivatedCommentsRow
	for _, row := range rows {
		if row.Idcomments == c.Idcomments {
			found = row
			break
		}
	}
	if found == nil {
		return fmt.Errorf("restore comment %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("not found")))
	}
	if err := cd.AdminRestoreComment(r.Context(), found.Idcomments, found.Text); err != nil {
		return fmt.Errorf("restore comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return nil
}
