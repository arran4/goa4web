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

type AdminCommentsPage struct{}

func (p *AdminCommentsPage) Action(w http.ResponseWriter, r *http.Request) any {
	return p
}

func (p *AdminCommentsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Comments"
	queries := cd.Queries()
	rows, err := queries.AdminListAllCommentsWithThreadInfo(r.Context(), db.AdminListAllCommentsWithThreadInfoParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	data := struct {
		*common.CoreData
		Comments []*db.AdminListAllCommentsWithThreadInfoRow
	}{cd, rows}
	AdminCommentsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminCommentsPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return "Comments", "/admin/comments", &AdminPage{}
}

func (p *AdminCommentsPage) PageTitle() string {
	return "Comments"
}

var _ tasks.Page = (*AdminCommentsPage)(nil)
var _ tasks.Task = (*AdminCommentsPage)(nil)
var _ http.Handler = (*AdminCommentsPage)(nil)

const AdminCommentsPageTmpl tasks.Template = "admin/commentsPage.gohtml"

type AdminCommentPage struct {
	CommentID int32
	Data      any
}

func (p *AdminCommentPage) Breadcrumb() (string, string, tasks.HasBreadcrumb) {
	return fmt.Sprintf("Comment %d", p.CommentID), "", &AdminCommentsPage{}
}

func (p *AdminCommentPage) PageTitle() string {
	return fmt.Sprintf("Comment %d", p.CommentID)
}

func (p *AdminCommentPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	AdminCommentPageTmpl.Handler(p.Data).ServeHTTP(w, r)
}

var _ tasks.Page = (*AdminCommentPage)(nil)
var _ http.Handler = (*AdminCommentPage)(nil)

type AdminCommentTask struct{}

func (t *AdminCommentTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		return handlers.ErrNotFound
	}

	queries := cd.Queries()
	rows, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: cd.UserID,
		Ids:      []int32{c.Idcomments},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil || len(rows) == 0 {
		return handlers.ErrNotFound
	}
	comment := rows[0]
	threadRows, _ := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: comment.ForumthreadID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	var contextRows []*db.GetCommentsByThreadIdForUserRow
	for i, row := range threadRows {
		if row.Idcomments == comment.Idcomments {
			start := i - 3
			if start < 0 {
				start = 0
			}
			end := i + 4
			if end > len(threadRows) {
				end = len(threadRows)
			}
			contextRows = threadRows[start:end]
			break
		}
	}
	data := struct {
		Comment *db.GetCommentsByIdsForUserWithThreadInfoRow
		Context []*db.GetCommentsByThreadIdForUserRow
	}{comment, contextRows}

	return &AdminCommentPage{
		CommentID: c.Idcomments,
		Data: data,
	}
}

const AdminCommentPageTmpl tasks.Template = "admin/adminCommentPage.gohtml"
