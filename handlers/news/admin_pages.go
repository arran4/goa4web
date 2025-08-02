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
	cd.PageTitle = "News Admin"

	posts, err := cd.LatestNewsList(0, 50)
	if err != nil {
		log.Printf("LatestNewsList: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*common.CoreData
		Posts []*common.NewsPost
	}{
		CoreData: cd,
		Posts:    posts,
	}
	handlers.TemplateHandler(w, r, "adminNewsListPage.gohtml", data)
}

type CommentPlus struct {
	*db.GetCommentsByThreadIdForUserRow
	ShowReply          bool
	EditUrl            string
	Editing            bool
	Offset             int
	Languages          []*db.Language
	SelectedLanguageId int32
	EditSaveUrl        string
}

func AdminNewsPostPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["post"])
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

	commentRows, err := queries.GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: int32(post.ForumthreadID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("GetCommentsByThreadIdForUser: %v", err)
	}
	var comments []*CommentPlus
	for i, row := range commentRows {
		comments = append(comments, &CommentPlus{
			GetCommentsByThreadIdForUserRow: row,
			ShowReply:                       false,
			EditUrl:                         "",
			EditSaveUrl:                     "",
			Editing:                         false,
			Offset:                          i,
		})
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
	data := struct {
		*common.CoreData
		Post     *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		TopicID  int32
		Thread   *db.GetThreadLastPosterAndPermsRow
		Comments []*CommentPlus
	}{
		CoreData: cd,
		Post:     post,
		TopicID:  topicID,
		Thread:   threadRow,
		Comments: comments,
	}
	handlers.TemplateHandler(w, r, "adminNewsPostPage.gohtml", data)
}

func adminNewsEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["post"])
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
	pid, err := strconv.Atoi(mux.Vars(r)["post"])
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
