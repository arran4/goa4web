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
	"github.com/gorilla/mux"
)

// adminCommentPage displays a single comment with nearby context.
func adminCommentPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = fmt.Sprintf("Comment %d", id)
	queries := cd.Queries()
	rows, err := queries.GetCommentsByIdsForUserWithThreadInfo(r.Context(), db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: cd.UserID,
		Ids:      []int32{int32(id)},
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
		*common.CoreData
		Comment *db.GetCommentsByIdsForUserWithThreadInfoRow
		Context []*db.GetCommentsByThreadIdForUserRow
	}{cd, comment, contextRows}
	handlers.TemplateHandler(w, r, "commentPage.gohtml", data)
}
