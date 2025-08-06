package news

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminNewsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/news?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/news?offset=%d", offset-ps)
		cd.StartLink = "/admin/news?offset=0"
	}
	cd.PageTitle = "News Admin"
	handlers.TemplateHandler(w, r, "adminNewsListPage.gohtml", struct{}{})
}

func AdminNewsPostPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Post           *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		TopicID        int32
		Thread         *db.GetThreadLastPosterAndPermsRow
		Comments       []*db.GetCommentsByThreadIdForUserRow
		IsReplyable    bool
		CanEditComment func(*db.GetCommentsByThreadIdForUserRow) bool
		EditURL        func(*db.GetCommentsByThreadIdForUserRow) string
		EditSaveURL    func(*db.GetCommentsByThreadIdForUserRow) string
		Editing        func(*db.GetCommentsByThreadIdForUserRow) bool
		AdminURL       func(*db.GetCommentsByThreadIdForUserRow) string
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusTemporaryRedirect)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		http.Redirect(w, r, "/admin/news?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	topicID, err := queries.GetForumTopicIdByThreadId(r.Context(), post.ForumthreadID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetForumTopicIdByThreadId: %v", err)
	}

	comments, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: int32(post.ForumthreadID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetCommentsByThreadIdForUser: %v", err)
	}
	threadRow, err := queries.GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
		ViewerID:      cd.UserID,
		ThreadID:      int32(post.ForumthreadID),
		ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetThreadLastPosterAndPerms: %v", err)
	}

	cd.PageTitle = fmt.Sprintf("News Post %d", pid)
	data := Data{
		CoreData:    cd,
		Post:        post,
		TopicID:     topicID,
		Thread:      threadRow,
		Comments:    comments,
		IsReplyable: false,
	}
	data.CanEditComment = func(*db.GetCommentsByThreadIdForUserRow) bool { return false }
	data.EditURL = func(*db.GetCommentsByThreadIdForUserRow) string { return "" }
	data.EditSaveURL = func(*db.GetCommentsByThreadIdForUserRow) string { return "" }
	data.Editing = func(*db.GetCommentsByThreadIdForUserRow) bool { return false }
	data.AdminURL = func(cmt *db.GetCommentsByThreadIdForUserRow) string {
		if cd.HasRole("administrator") {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}

	handlers.TemplateHandler(w, r, "adminNewsPostPage.gohtml", data)
}

func adminNewsEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusTemporaryRedirect)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		http.Redirect(w, r, "/admin/news?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	langs, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	cd.PageTitle = "Edit News"
	data := struct {
		*common.CoreData
		Languages          []*db.Language
		Post               *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		SelectedLanguageId int
	}{
		CoreData:           cd,
		Languages:          langs,
		Post:               post,
		SelectedLanguageId: int(post.LanguageIdlanguage),
	}
	handlers.TemplateHandler(w, r, "adminNewsEditPage.gohtml", data)
}

func AdminNewsDeleteConfirmPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusTemporaryRedirect)
		return
	}
	cd.PageTitle = "Confirm news delete"
	data := struct {
		*common.CoreData
		PostID       int
		ConfirmLabel string
		Back         string
	}{
		CoreData:     cd,
		PostID:       pid,
		ConfirmLabel: "Confirm delete",
		Back:         fmt.Sprintf("/admin/news/%d", pid),
	}
	handlers.TemplateHandler(w, r, "adminNewsDeleteConfirmPage.gohtml", data)
}

// NewsDelete deactivates a news post.
func NewsDelete(ctx context.Context, q db.Querier, postID int32) error {
	return q.DeactivateNewsPost(ctx, postID)
}
