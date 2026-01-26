package admin

import (
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

// adminCommentPage displays a single comment with nearby context.
func adminCommentPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	c, err := cd.CurrentComment(r)
	if err != nil || c == nil {
		http.NotFound(w, r)
		return
	}
	cd.PageTitle = fmt.Sprintf("Comment %d", c.Idcomments)
	queries := cd.Queries()
	rows, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: cd.UserID,
		Ids:      []int32{c.Idcomments},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil || len(rows) == 0 {
		http.NotFound(w, r)
		return
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
	AdminCommentPageTmpl.Handle(w, r, data)
}

const AdminCommentPageTmpl tasks.Template = "admin/adminCommentPage.gohtml"
