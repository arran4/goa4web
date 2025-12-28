package news

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/admincommon"
	"github.com/arran4/goa4web/internal/db"
)

func AdminNewsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Error     string
		CanPost   bool
		UserRoles []admincommon.UserRoleInfo
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	ps := cd.PageSize()
	cd.NextLink = fmt.Sprintf("/admin/news?offset=%d", offset+ps)
	if offset > 0 {
		cd.PrevLink = fmt.Sprintf("/admin/news?offset=%d", offset-ps)
		cd.StartLink = "/admin/news?offset=0"
	}
	cd.PageTitle = "News Admin"

	data := Data{Error: r.URL.Query().Get("error"), CanPost: cd.HasGrant("news", "post", "edit", 0) && cd.AdminMode}
	queries := cd.Queries()
	userRoles, err := admincommon.LoadUserRoleInfo(r.Context(), queries, nil)
	if err == nil {
		data.UserRoles = userRoles
	}
	sort.Slice(data.UserRoles, func(i, j int) bool {
		return data.UserRoles[i].Username.String < data.UserRoles[j].Username.String
	})

	if err := cd.ExecuteSiteTemplate(w, r, "news/adminNewsListPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func AdminNewsPostPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
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
		http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "/admin/news", err)
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
		if cd.HasAdminRole() {
			return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
		}
		return ""
	}

	if err := cd.ExecuteSiteTemplate(w, r, "news/adminNewsPostPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func adminNewsEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
		return
	}
	post, err := queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
		ViewerID: cd.UserID,
		ID:       int32(pid),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		handlers.RedirectSeeOtherWithError(w, r, "/admin/news", err)
		return
	}
	langs, err := cd.Languages()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}
	cd.PageTitle = "Edit News"
	labels, _ := cd.NewsAuthorLabels(post.Idsitenews)
	data := struct {
		Languages          []*db.Language
		Post               *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
		SelectedLanguageId int
		AuthorLabels       []string
	}{
		Languages:          langs,
		Post:               post,
		SelectedLanguageId: int(post.LanguageID.Int32),
		AuthorLabels:       labels,
	}
	if err := cd.ExecuteSiteTemplate(w, r, "news/adminNewsEditPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}

func AdminNewsDeleteConfirmPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	pid, err := strconv.Atoi(mux.Vars(r)["news"])
	if err != nil {
		http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
		return
	}
	cd.PageTitle = "Confirm news delete"
	data := struct {
		PostID       int
		ConfirmLabel string
		Back         string
	}{
		PostID:       pid,
		ConfirmLabel: "Confirm delete",
		Back:         fmt.Sprintf("/admin/news/article/%d", pid),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "news/adminNewsDeleteConfirmPage.gohtml", data); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
